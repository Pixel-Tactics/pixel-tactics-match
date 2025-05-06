package communication

import (
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/gateway"
)

type IncomingQueue struct {
	Router       gateway.Router
	EventManager events.EventManager
	RMQManager   *RMQManager
}

func NewIncomingQueue(
	rmqManager *RMQManager,
) *IncomingQueue {
	queue := &IncomingQueue{
		RMQManager: rmqManager,
	}
	return queue
}

func (queue *IncomingQueue) Run(queueName string, foo func(data interface{}) error) {
	retrying := false
	for {
		if retrying {
			log.Println("channel closed, retrying...")
			time.Sleep(3 * time.Second)
		}
		retrying = true

		log.Println("[INFO] Connecting to queue " + queueName + "...")

		channelId := INCOMING_CHANNEL + "invite:" + queueName
		channel, err := queue.RMQManager.GetChannel(channelId)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = channel.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			amqp091.Table{
				// "x-queue-type": "stream",
			},
		)
		if err != nil {
			log.Println(err)
			continue
		}

		queue.consume(channel, queueName, foo)
	}
}

func (queue *IncomingQueue) consume(channel *amqp091.Channel, queueName string, foo func(data interface{}) error) {
	consumer, err := channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Println(err)
		return
	}
	for delivery := range consumer {
		log.Println(string(delivery.Body))
		err := foo(delivery.Body)
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
	log.Println("[WARNING] Disconnected from queue " + queueName)
}
