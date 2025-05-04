package events

type EventManager interface {
	// Emits event to key
	Emit(key string, value interface{}) error

	// Listens to event. Returns removal id.
	On(key string, foo func(interface{}) error) string

	// Unlistens to event using id returned from `On()` function.
	RemoveOn(key string, id string)
}
