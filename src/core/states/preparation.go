package states

import (
	"time"

	"github.com/google/uuid"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/models"
)

type PreparationState struct {
	Session      *models.Session
	EventManager events.EventManager
}

func (state *PreparationState) Start(deadline time.Time) error {
	state.Session.State = models.State{
		Id:       uuid.New().String(),
		Type:     models.SessionStatePlayer1Turn,
		Deadline: deadline,
	}
	// TODO: move state update event to be in channel to handle kafka fails
	return sendStateUpdateEvent(state.EventManager, state.Session.Id, state.Session.State)
}

func (state *PreparationState) Swap(deadline time.Time) error {
	return exceptions.ActionNotAllowed()
}

func (state *PreparationState) End(winnerId *string) error {
	state.Session.WinnerId = winnerId
	state.Session.State = models.State{
		Id:   uuid.New().String(),
		Type: models.SessionStateEnded,
	}
	return sendStateUpdateEvent(state.EventManager, state.Session.Id, state.Session.State)
}

func NewPreparationState(session *models.Session, eventManager events.EventManager) *PreparationState {
	return &PreparationState{
		Session:      session,
		EventManager: eventManager,
	}
}
