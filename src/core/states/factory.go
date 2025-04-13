package states

import (
	"pixeltactics.com/match/src/models"
)

type SessionStateFactory interface {
	Create(session *models.Session) SessionState
}

type SessionStateFactoryImpl struct{}

func (factory *SessionStateFactoryImpl) Create(session *models.Session) SessionState {
	if session.State.Type == models.SessionStatePreparation {
		return NewPreparationState(session)
	} else if session.State.Type == models.SessionStatePlayer1Turn || session.State.Type == models.SessionStatePlayer2Turn {
		return NewPlayerTurnState(session)
	} else if session.State.Type == models.SessionStateEnded {
		return NewEndState()
	} else {
		return NewMatchmakingState(session)
	}
}

func NewSessionStateFactory() SessionStateFactory {
	return &SessionStateFactoryImpl{}
}
