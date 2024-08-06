package notifiers

import (
	"log"

	ws_types "pixeltactics.com/match/src/websocket/types"
)

type SessionNotifier struct {
	SendChannel         chan *NotifierData
	sendMessageToPlayer func(string, *ws_types.Message)
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

func (notifier *SessionNotifier) Run() {
	for {
		msg, ok := <-notifier.SendChannel
		if ok {
			playerId := msg.PlayerId
			message := msg.Message
			notifier.sendMessageToPlayer(playerId, &message)
		}
	}
}

var sessionNotifier *SessionNotifier = nil

func InitSessionNotifier(sendMessageToPlayer func(string, *ws_types.Message)) {
	sessionNotifier = &SessionNotifier{
		SendChannel:         make(chan *NotifierData, 256),
		sendMessageToPlayer: sendMessageToPlayer,
	}
}

func GetSessionNotifier() *SessionNotifier {
	if sessionNotifier == nil {
		log.Fatal("Session Notifier is nil")
	}
	return sessionNotifier
}
