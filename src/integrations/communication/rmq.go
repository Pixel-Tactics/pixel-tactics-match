package communication

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"pixeltactics.com/match/src/config"
)

const (
	RECONNECT_COOLDOWN = 5 * time.Second
)

var ErrReconnectFailed = errors.New("reconnect failed")

type RMQManager struct {
	Conn *amqp091.Connection

	LastConnectTime time.Time
	Channels        map[string]*amqp091.Channel

	*sync.RWMutex
}

func NewRMQManager() *RMQManager {
	obj := &RMQManager{
		Channels: make(map[string]*amqp091.Channel),
		RWMutex:  new(sync.RWMutex),
	}
	err := obj.connect()
	if err != nil {
		log.Println("cannot do initial connect")
	}
	return obj
}

func (rmq *RMQManager) isConnected() bool {
	return rmq.Conn != nil && !rmq.Conn.IsClosed()
}

func (rmq *RMQManager) connect() error {
	if time.Since(rmq.LastConnectTime) < RECONNECT_COOLDOWN {
		return ErrReconnectFailed
	}

	conn, err := amqp091.Dial(config.RMQUrl)
	if err != nil {
		log.Println(err)
		return err
	}
	rmq.Conn = conn
	rmq.LastConnectTime = time.Now()
	return nil
}

// Gets channel with specified id. Do not forget to close the channel, see `CloseChannel(id string)`.
// The function is thread-safe, but the channels are not. So, make sure only one thread accesses one id.
func (rmq *RMQManager) GetChannel(id string) (*amqp091.Channel, error) {
	rmq.RLock()
	channel, ok := rmq.Channels[id]
	if ok && !channel.IsClosed() {
		rmq.RUnlock()
		return channel, nil
	}

	rmq.RUnlock()
	rmq.Lock()
	defer rmq.Unlock()

	if !rmq.isConnected() {
		err := rmq.connect()
		if err != nil {
			return nil, err
		}
	}
	channel, err := rmq.Conn.Channel()
	if err != nil {
		return nil, err
	}
	rmq.Channels[id] = channel
	return channel, nil
}

// Closes channel with specified id.
// The function is thread-safe, but the channels are not. So, make sure only one thread accesses one id.
func (rmq *RMQManager) CloseChannel(id string) {
	rmq.Lock()
	defer rmq.Unlock()

	channel, ok := rmq.Channels[id]
	if ok {
		channel.Close()
	}
	delete(rmq.Channels, id)
}
