package states

import (
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/models"
)

const STATE_UPDATE_EVENT = "session_state_update"

type StateUpdateEvent struct {
	SessionId string
	PlayerIDs []string
	NewState  models.State
}

func sendStateUpdateEvent(manager events.EventManager, session *models.Session) error {
	return manager.Emit(STATE_UPDATE_EVENT, &StateUpdateEvent{
		SessionId: session.Id,
		PlayerIDs: session.PlayerIds,
		NewState:  session.State,
	})
}
