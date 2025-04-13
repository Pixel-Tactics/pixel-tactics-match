package states

import (
	"log"
	"time"

	"github.com/google/uuid"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/models"
)

type MatchmakingState struct {
	Session *models.Session
}

func (state *MatchmakingState) Start(deadline time.Time) error {
	log.Println("SETTING TO PREPARATION")
	state.Session.State = models.State{
		Id:       uuid.New().String(),
		Type:     models.SessionStatePreparation,
		Deadline: deadline,
	}
	return nil
}

func (state *MatchmakingState) Swap(deadline time.Time) error {
	return exceptions.ActionNotAllowed()
}

func (state *MatchmakingState) End(winnerId *string) error {
	return exceptions.ActionNotAllowed()
}

func NewMatchmakingState(session *models.Session) *MatchmakingState {
	return &MatchmakingState{
		Session: session,
	}
}
