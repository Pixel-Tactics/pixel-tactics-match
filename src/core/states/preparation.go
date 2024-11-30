package states

import (
	"time"

	"github.com/google/uuid"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/models"
)

type PreparationState struct {
	session *models.Session
}

func (state *PreparationState) Start(deadline time.Time) error {
	state.session.State = models.State{
		Id:       uuid.New().String(),
		Type:     models.SessionStatePlayer1Turn,
		Deadline: deadline,
	}
	return nil
}

func (state *PreparationState) Swap(deadline time.Time) error {
	return exceptions.ActionNotAllowed()
}

func (state *PreparationState) End(winnerId *string) error {
	state.session.WinnerId = winnerId
	state.session.State = models.State{
		Id:   uuid.New().String(),
		Type: models.SessionStateEnded,
	}
	return nil
}

func NewPreparationState(session *models.Session) *PreparationState {
	return &PreparationState{
		session: session,
	}
}
