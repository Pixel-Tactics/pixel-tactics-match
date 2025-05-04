package communication

import (
	"encoding/json"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/gateway"
	"pixeltactics.com/match/src/messages"
)

const SEND_EVENT = "send:"
const INCOMING_CHANNEL = "incoming:"
const INCOMING_PREFIX = "session/incoming/"

type IncomingQueue struct {
	Router       gateway.Router
	EventManager events.EventManager
	RMQManager   *RMQManager

	Users map[string]*UserQueue

	AddUser    chan string
	DeleteUser chan string
}

func NewIncomingQueue(
	router gateway.Router,
	rmqManager *RMQManager,
	eventManager events.EventManager,
) *IncomingQueue {
	queue := &IncomingQueue{
		Router:     router,
		RMQManager: rmqManager,
		Users:      make(map[string]*UserQueue),
		AddUser:    make(chan string, 256),
		DeleteUser: make(chan string, 256),
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

func (queue *IncomingQueue) Run() {
	for {
		select {
		case username := <-queue.AddUser:
			log.Println("Got add user")
			_, exists := queue.Users[username]
			if exists {
				continue
			}
			userQueue := NewUserQueue(username, queue, queue.Router, queue.EventManager)
			queue.Users[username] = userQueue
			go userQueue.Run()
		case username := <-queue.DeleteUser:
			log.Println("Got delete user")
			userQueue, exists := queue.Users[username]
			if !exists {
				continue
			}
			userQueue.Close <- true
			delete(queue.Users, username)
		}
	}
}

type UserQueue struct {
	Username string
	Parent   *IncomingQueue
	Messages chan *messages.Message
	Close    chan bool

	Router       gateway.Router
	EventManager events.EventManager
}

func NewUserQueue(username string, parent *IncomingQueue, router gateway.Router, eventManager events.EventManager) *UserQueue {
	return &UserQueue{
		Username:     username,
		Parent:       parent,
		Messages:     make(chan *messages.Message, 256),
		Close:        make(chan bool, 256),
		Router:       router,
		EventManager: eventManager,
	}
}

// Runs the user queue.
func (queue *UserQueue) Run() {
	retrying := false
	for {
		if retrying {
			log.Println("channel closed, retrying...")
			time.Sleep(3 * time.Second)
		}
		retrying = true

		channelId := INCOMING_CHANNEL + "session" + ":" + queue.Username
		channel, err := queue.Parent.RMQManager.GetChannel(channelId)
		if err != nil {
			log.Println(err)
			continue
		}

		queueName := INCOMING_PREFIX + queue.Username
		_, err = channel.QueueDeclare(
			queueName,
			true,
			false,
			false,
			false,
			amqp091.Table{
				"x-expires": int32(60000),
			},
		)
		if err != nil {
			log.Println(err)
			continue
		}

		stop := queue.consume(channelId, channel, queueName)
		if stop {
			break
		}
	}
}

// Listens for RabbitMQ queue and close signal. Returns false if lost connection, or true if intended to be closed.
func (queue *UserQueue) consume(channelId string, channel *amqp091.Channel, queueName string) bool {
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
		return false
	}
	for {
		select {
		case <-queue.Close:
			queue.Parent.RMQManager.CloseChannel(channelId)
			return true
		case delivery, ok := <-consumer:
			if !ok {
				return false
			}

			var message *messages.Message
			errBody := json.Unmarshal(delivery.Body, &message)
			err = delivery.Ack(false)
			if err != nil {
				return false
			}
			if errBody != nil {
				log.Println(errBody)
				continue
			}

			queue.Router.RouteMessage(queue, message)
		}
	}
}

func (queue *UserQueue) GetUsername() string {
	return queue.Username
}

func (queue *UserQueue) Send(username string, message *messages.Message) {
	queue.EventManager.Emit(SEND_EVENT+username, message)
}

func (queue *UserQueue) SendBack(message *messages.Message) {
	queue.EventManager.Emit(SEND_EVENT+queue.Username, message)
}
