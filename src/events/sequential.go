package events

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

type SequentialEventManager struct {
	Subscribers map[string]map[string]func(interface{}) error
	*sync.RWMutex
}

func (manager *SequentialEventManager) Emit(key string, value interface{}) error {
	manager.RLock()
	defer manager.RUnlock()

	subscribers, ok := manager.Subscribers[key]
	if !ok {
		return nil
	}
	hasError := false
	for _, subscriber := range subscribers {
		err := subscriber(value)
		if err != nil {
			hasError = true
		}
	}
	if hasError {
		return errors.New("subscriber error")
	} else {
		return nil
	}
}

func (manager *SequentialEventManager) On(key string, foo func(interface{}) error) string {
	manager.Lock()
	defer manager.Unlock()

	subscribers, ok := manager.Subscribers[key]
	if !ok {
		subscribers = make(map[string]func(interface{}) error)
	}
	id := uuid.New().String()
	subscribers[id] = foo
	manager.Subscribers[key] = subscribers
	return id
}

func (manager *SequentialEventManager) RemoveOn(key string, id string) {
	manager.Lock()
	defer manager.Unlock()

	subscribers, ok := manager.Subscribers[key]
	if !ok {
		return
	}
	delete(subscribers, id)
}

func (manager *SequentialEventManager) Run() error {
	panic("Invalid method")
}

func NewSequentialEventManager() *SequentialEventManager {
	return &SequentialEventManager{
		Subscribers: make(map[string]map[string]func(interface{}) error),
		RWMutex:     new(sync.RWMutex),
	}
}
