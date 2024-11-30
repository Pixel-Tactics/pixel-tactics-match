package states

import "time"

type SessionState interface {
	Start(deadline time.Time) error
	Swap(deadline time.Time) error
	End(winnerId *string) error
}

// type TimedState interface {
// 	GetDeadline() time.Time
// 	Expire()
// }
