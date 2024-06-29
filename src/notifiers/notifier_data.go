package notifiers

import ws_types "pixeltactics.com/match/src/websocket/types"

type NotifierData struct {
	PlayerId string
	Message  ws_types.Message
}
