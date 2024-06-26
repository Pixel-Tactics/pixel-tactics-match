package notifiers

import (
	"sync"

	ws_types "pixeltactics.com/match/src/websocket/types"
)

type SessionNotifier struct {
	SendChannel chan *NotifierData
}

func (notifier *SessionNotifier) NotifyChangeState(playerId string, sessionData map[string]interface{}) {
	notifier.SendChannel <- &NotifierData{
		PlayerId: playerId,
		Message: ws_types.Message{
			Action: "STATE_CHANGE",
			Body: map[string]interface{}{
				"session": sessionData,
			},
		},
	}
}

func (notifier *SessionNotifier) NotifyAction(playerId string, actionName string, actionData map[string]interface{}) {
	notifier.SendChannel <- &NotifierData{
		PlayerId: playerId,
		Message: ws_types.Message{
			Action: ws_types.ACTION_APPLY_ACTION,
			Body: map[string]interface{}{
				"actionName":     actionName,
				"actionSpecific": actionData,
			},
		},
	}
}

var sessionNotifier *SessionNotifier = nil
var once sync.Once

func GetSessionNotifier() *SessionNotifier {
	once.Do(func() {
		sessionNotifier = &SessionNotifier{
			SendChannel: make(chan *NotifierData, 256),
		}
	})
	return sessionNotifier
}
