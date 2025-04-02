package events

type EventManager interface {
	Emit(key string, value interface{}) error
	On(key string, foo func(interface{}) error)
}
