package communication

import (
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/gateway"
)

type IncomingStream struct {
	Router       gateway.Router
	EventManager events.EventManager
	RMQManager   *RMQManager
}

func NewIncomingStream(
	rmqManager *RMQManager,
	eventManager events.EventManager,
) *IncomingStream {
	queue := &IncomingStream{
		EventManager: eventManager,
		RMQManager:   rmqManager,
	}
	return queue
}

func (stream *IncomingStream) Run(streamName string, eventName string) {
	retrying := false
	for {
		if retrying {
			log.Println("channel closed, retrying...")
			time.Sleep(3 * time.Second)
		}
		retrying = true

		log.Println("[INFO] Connecting to stream " + streamName + "...")

		channelId := INCOMING_CHANNEL + "invite:" + streamName
		channel, err := stream.RMQManager.GetChannel(channelId)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = channel.QueueDeclare(
			streamName,
			true,
			false,
			false,
			false,
			amqp091.Table{
				"x-queue-type": "stream",
			},
		)
		if err != nil {
			log.Println(err)
			continue
		}

		stream.consume(channel, streamName, eventName)
	}
}

func (stream *IncomingStream) consume(channel *amqp091.Channel, streamName string, eventName string) {
	err := channel.Qos(
		100,   // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Println(err)
		return
	}

	consumer, err := channel.Consume(
		streamName,
		"consumer-"+streamName, // add a consumer name for offset tracking
		false,                  // auto-ack
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		amqp091.Table{
			"x-stream-offset": "last", // start from beginning if no offset
		},
	)
	if err != nil {
		log.Println(err)
		return
	}
	for delivery := range consumer {
		log.Println(string(delivery.Body))
		err := stream.EventManager.Emit(eventName, delivery.Body)
		if err != nil {
			log.Println(err)
			continue // TODO: masukin dead letter queue
		}
		err = delivery.Ack(false) // acknowledge the message after processing
		if err != nil {
			log.Println(err)
			return
		}
	}
	log.Println("[WARNING] Disconnected from stream " + streamName)
}
