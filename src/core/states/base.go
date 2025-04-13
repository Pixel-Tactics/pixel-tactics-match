package states

import (
	"time"

	"github.com/google/uuid"
	"pixeltactics.com/match/src/models"
)

type SessionState interface {
	Start(deadline time.Time) error
	Swap(deadline time.Time) error
	End(winnerId *string) error
}

type BaseSessionState struct {
	Session *models.Session
}

func (state *BaseSessionState) End(winnerId *string) error {
	state.Session.WinnerId = winnerId
	state.Session.State = models.State{
		Id:   uuid.New().String(),
		Type: models.SessionStateEnded,
	}
	return nil
}
