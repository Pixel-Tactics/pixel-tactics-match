package states

import (
	"time"

	"github.com/google/uuid"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/models"
)

type PlayerTurnState struct {
	*BaseSessionState
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
	state.Session.CurrentTurn += 1
	return nil
}

func (state *PlayerTurnState) GetActivePlayerID() string {
	if state.Session.State.Type == models.SessionStatePlayer1Turn {
		return state.Session.PlayerIds[0]
	}
	return state.Session.PlayerIds[1]
}

func (state *PlayerTurnState) GetInactivePlayerID() string {
	if state.Session.State.Type == models.SessionStatePlayer1Turn {
		return state.Session.PlayerIds[1]
	}
	return state.Session.PlayerIds[0]
}

func NewPlayerTurnState(session *models.Session) SessionState {
	return &PlayerTurnState{
		BaseSessionState: &BaseSessionState{
			Session: session,
		},
	}
}
