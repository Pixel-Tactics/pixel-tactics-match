package gateway

import "pixeltactics.com/match/src/messages"

func Error(err error) *messages.Message {
	return &messages.Message{
		Type: "ERROR",
		Body: map[string]interface{}{
			"message": err.Error(),
		},
	}
}
