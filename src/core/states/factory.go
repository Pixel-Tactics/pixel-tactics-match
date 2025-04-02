package states

import (
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/models"
)

type SessionStateFactory interface {
	Create(session *models.Session) SessionState
}

type SessionStateFactoryImpl struct {
	EventManager events.EventManager
}

func (factory *SessionStateFactoryImpl) Create(session *models.Session) SessionState {
	if session.State.Type == models.SessionStatePreparation {
		return NewPreparationState(session, factory.EventManager)
	} else if session.State.Type == models.SessionStatePlayer1Turn || session.State.Type == models.SessionStatePlayer2Turn {
		return NewPlayerTurnState(session, factory.EventManager)
	} else if session.State.Type == models.SessionStateEnded {
		return NewEndState(factory.EventManager)
	} else {
		return NewMatchmakingState(session, factory.EventManager)
	}
}

func NewSessionStateFactory(eventManager events.EventManager) SessionStateFactory {
	return &SessionStateFactoryImpl{
		EventManager: eventManager,
	}
}
