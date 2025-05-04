package gateway

import "pixeltactics.com/match/src/messages"

func Error(err error) *messages.Message {
	return &messages.Message{
		Route: "ERROR",
		Data: map[string]interface{}{
			"message": err.Error(),
		},
	}
}
