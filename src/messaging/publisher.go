package messaging

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn     *amqp.Connection
	chann    *amqp.Channel
	waitList chan *PublisherMessage
}

func (pub *Publisher) Publish(pubMsg *PublisherMessage) {
	pub.waitList <- pubMsg
}

func (pub *Publisher) ensureConnection() {
	for {
		if pub.chann == nil || pub.conn == nil || pub.chann.IsClosed() || pub.conn.IsClosed() {
			conn, chann, err := createConnection()
			if err != nil {
				log.Println("Error on connecting to RabbitMQ Server, retrying in 5 seconds..")
				time.Sleep(5 * time.Second)
				continue
			}
			pub.conn = conn
			pub.chann = chann
		} else {
			break
		}
	}
}

func (pub *Publisher) close() {
	pub.conn.Close()
	pub.chann.Close()
}

func createConnection() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONNECTION_STRING"))
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	return conn, ch, nil
}

func newPublisher() *Publisher {
	publisher := &Publisher{
		conn:     nil,
		chann:    nil,
		waitList: make(chan *PublisherMessage, 256),
	}
	go func() {
		defer func() {
			publisher.close()
		}()

		for {
			msg, ok := <-publisher.waitList
			if ok {
				publisher.ensureConnection()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

				err := publisher.chann.PublishWithContext(
					ctx,
					msg.Exchange,
					msg.RoutingKey,
					true,
					false,
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(msg.Body),
					},
				)
				if err != nil {
					publisher.waitList <- msg
				}

				cancel()
			}
		}
	}()
	return publisher
}

var publisher *Publisher = nil
var lock *sync.Mutex = new(sync.Mutex)

func GetPublisher() *Publisher {
	lock.Lock()
	defer lock.Unlock()
	if publisher == nil {
		publisher = newPublisher()
	}
	return publisher
}
