package states

import (
	"time"

	"github.com/google/uuid"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/models"
)

type PreparationState struct {
	*BaseSessionState
}

func (state *PreparationState) Start(deadline time.Time) error {
	state.Session.State = models.State{
		Id:       uuid.New().String(),
		Type:     models.SessionStatePlayer1Turn,
		Deadline: deadline,
	}
	return nil
}

func (state *PreparationState) Swap(deadline time.Time) error {
	return exceptions.ActionNotAllowed()
}

func NewPreparationState(session *models.Session) SessionState {
	return &PreparationState{
		BaseSessionState: &BaseSessionState{
			Session: session,
		},
	}
}
