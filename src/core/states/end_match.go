package states

import (
	"time"

	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/models"
)

type EndState struct {
	*BaseSessionState
}

func (state *EndState) Start(deadline time.Time) error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) Swap(deadline time.Time) error {
	return exceptions.ActionNotAllowed()
}

func NewEndState(session *models.Session) SessionState {
	return &EndState{
		BaseSessionState: &BaseSessionState{
			Session: session,
		},
	}
}
