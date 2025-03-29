package events

import "errors"

type EventManager interface {
	Emit(key string, value map[string]interface{}) error
	On(key string, foo func(map[string]interface{}) error)
}

type SequentialEventManager struct {
	Subscribers map[string][]func(map[string]interface{}) error
}

type Event struct {
	Key   string
	Value map[string]interface{}
}

func (manager *SequentialEventManager) Emit(key string, value map[string]interface{}) error {
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

func (manager *SequentialEventManager) On(key string, foo func(map[string]interface{}) error) {
	subscribers, ok := manager.Subscribers[key]
	if !ok {
		subscribers = make([]func(map[string]interface{}) error, 0)
	}
	subscribers = append(subscribers, foo)
	manager.Subscribers[key] = subscribers
}

func (manager *SequentialEventManager) Run() error {
	panic("Invalid method")
}

// type EventManagerImpl struct {
// 	Channel     chan *Event
// 	Subscribers map[string][]func(string, map[string]interface{})
// }

// type Event struct {
// 	Key   string
// 	Value map[string]interface{}
// }

// func (manager *EventManagerImpl) Emit(key string, value map[string]interface{}) {
// 	manager.Channel <- &Event{
// 		Key:   key,
// 		Value: value,
// 	}
// }

// func (manager *EventManagerImpl) On(key string, foo func(string, map[string]interface{})) {
// 	subscribers, ok := manager.Subscribers[key]
// 	if !ok {
// 		panic("no event called " + key)
// 	}
// 	subscribers = append(subscribers, foo)
// 	manager.Subscribers[key] = subscribers
// }

// func (manager *EventManagerImpl) Run() {
// 	for event := range manager.Channel {
// 		subscribers, ok := manager.Subscribers[event.Key]
// 		if !ok {
// 			continue
// 		}
// 		for _, foo := range subscribers {
// 			foo(event.Key, event.Value)
// 		}
// 	}
// }
