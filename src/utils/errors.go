package utils

import (
	ws_types "pixeltactics.com/match/src/websocket/types"
)

func ErrorMessage(err error) *ws_types.Message {
	return &ws_types.Message{
		Action: ws_types.ACTION_ERROR,
		Body: map[string]interface{}{
			"error": err.Error(),
		},
	}
}
