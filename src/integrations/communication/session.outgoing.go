package communication

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/messages"
)

const OUTGOING_CHANNEL = "outgoing"
const OUTGOING_PREFIX = "session/outgoing/"

type OutgoingQueue struct {
	EventManager events.EventManager
	RMQManager   *RMQManager

	SubscribeIds map[string]string // username to subscribe id

	AddUser    chan string
	DeleteUser chan string
	Messages   chan UserMessage
}

type UserMessage struct {
	Username string
	Message  *messages.Message
}

func NewOutgoingQueue(
	rmqManager *RMQManager,
	eventManager events.EventManager,
) *OutgoingQueue {
	queue := &OutgoingQueue{
		RMQManager:   rmqManager,
		SubscribeIds: make(map[string]string),
		AddUser:      make(chan string, 256),
		DeleteUser:   make(chan string, 256),
		Messages:     make(chan UserMessage, 256),
	}
	eventManager.On("user-added", func(event interface{}) error {
		username, ok := event.(string)
		if !ok {
			log.Println("[ERROR] cannot convert event to string")
			return nil
		}
		queue.AddUser <- username
		return nil
	})
	eventManager.On("user-removed", func(event interface{}) error {
		username, ok := event.(string)
		if !ok {
			log.Println("[ERROR] cannot convert event to string")
			return nil
		}
		queue.DeleteUser <- username
		return nil
	})
	return queue
}

func (queue *OutgoingQueue) Run() {
	for {
		select {
		case username := <-queue.AddUser:
			log.Println("Got add user")
			subscribeId := queue.EventManager.On(SEND_EVENT+username, func(i interface{}) error {
				message, ok := i.(*messages.Message)
				if !ok {
					log.Println("[UNEXPECTED] invalid message from send event")
					log.Println(message)
				}
				queue.Messages <- UserMessage{
					Username: username,
					Message:  message,
				}
				return nil
			})
			_, ok := queue.SubscribeIds[username]
			if ok {
				log.Println("[UNEXPECTED] username already exists")
			}
			queue.SubscribeIds[username] = subscribeId
		case username := <-queue.DeleteUser:
			log.Println("Got delete user")
			subscribeId, ok := queue.SubscribeIds[username]
			if !ok {
				log.Println("[UNEXPECTED] subscribe id is missing")
			}
			queue.EventManager.RemoveOn(SEND_EVENT+username, subscribeId)
		case pair := <-queue.Messages:
			channel, err := queue.RMQManager.GetChannel(OUTGOING_CHANNEL)
			if err != nil {
				log.Println(err)
				continue
			}

			data, err := json.Marshal(pair.Message)
			if err != nil {
				log.Println(err)
				continue
			}

			channel.Publish("", OUTGOING_PREFIX+pair.Username, true, false, amqp091.Publishing{
				Body: data,
			})
		}
	}
}

func (queue *OutgoingQueue) Send(username string, message *messages.Message) {
	queue.Messages <- UserMessage{
		Username: username,
		Message:  message,
	}
}
