package events

import "errors"

type SequentialEventManager struct {
	Subscribers map[string][]func(interface{}) error
}

func (manager *SequentialEventManager) Emit(key string, value interface{}) error {
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

func (manager *SequentialEventManager) On(key string, foo func(interface{}) error) {
	subscribers, ok := manager.Subscribers[key]
	if !ok {
		subscribers = make([]func(interface{}) error, 0)
	}
	subscribers = append(subscribers, foo)
	manager.Subscribers[key] = subscribers
}

func (manager *SequentialEventManager) Run() error {
	panic("Invalid method")
}

func NewSequentialEventManager() *SequentialEventManager {
	return &SequentialEventManager{
		Subscribers: make(map[string][]func(interface{}) error),
	}
}
