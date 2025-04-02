package states

import (
	"time"

	"github.com/google/uuid"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/models"
)

type PlayerTurnState struct {
	Session      *models.Session
	EventManager events.EventManager
}

func (state *PlayerTurnState) Start(deadline time.Time) error {
	return exceptions.ActionNotAllowed()
}

func (state *PlayerTurnState) Swap(deadline time.Time) error {
	if state.Session.State.Type == models.SessionStatePlayer1Turn {
		state.Session.State = models.State{
			Id:       uuid.New().String(),
			Type:     models.SessionStatePlayer2Turn,
			Deadline: deadline,
		}
	} else {
		state.Session.State = models.State{
			Id:       uuid.New().String(),
			Type:     models.SessionStatePlayer1Turn,
			Deadline: deadline,
		}
	}
	return sendStateUpdateEvent(state.EventManager, state.Session.Id, state.Session.State)
}

func (state *PlayerTurnState) End(winnerId *string) error {
	return exceptions.ActionNotAllowed()
}

func NewPlayerTurnState(session *models.Session, eventManager events.EventManager) *PlayerTurnState {
	return &PlayerTurnState{
		Session:      session,
		EventManager: eventManager,
	}
}
