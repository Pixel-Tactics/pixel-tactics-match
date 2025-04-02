package events

type AsyncEventManager struct {
	Channel     chan *Event
	Subscribers map[string][]func(interface{})
}

type Event struct {
	Key   string
	Value interface{}
}

func (manager *AsyncEventManager) Emit(key string, value interface{}) error {
	manager.Channel <- &Event{
		Key:   key,
		Value: value,
	}
	return nil
}

func (manager *AsyncEventManager) On(key string, foo func(interface{})) {
	subscribers, ok := manager.Subscribers[key]
	if !ok {
		panic("no event called " + key)
	}
	subscribers = append(subscribers, foo)
	manager.Subscribers[key] = subscribers
}

func (manager *AsyncEventManager) Run() {
	for event := range manager.Channel {
		subscribers, ok := manager.Subscribers[event.Key]
		if !ok {
			continue
		}
		for _, foo := range subscribers {
			foo(event.Value)
		}
	}
}
