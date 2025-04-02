package states

import (
	"time"

	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/exceptions"
)

type EndState struct {
	EventManager events.EventManager
}

func (state *EndState) Start(deadline time.Time) error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) Swap(deadline time.Time) error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) End(winnerId *string) error {
	return exceptions.ActionNotAllowed()
}

func NewEndState(eventManager events.EventManager) *EndState {
	return &EndState{
		EventManager: eventManager,
	}
}
