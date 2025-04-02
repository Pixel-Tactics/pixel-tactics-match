package states

import (
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/models"
	convert_utils "pixeltactics.com/match/src/utils/convert"
)

func sendStateUpdateEvent(manager events.EventManager, sessionId string, obj models.State) error {
	mapObj, err := convert_utils.ObjectToMap(obj)
	if err != nil {
		return err
	}
	mapObj["sessionId"] = sessionId
	return manager.Emit("session_state_update", mapObj)
}
