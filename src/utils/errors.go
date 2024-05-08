package utils

import "pixeltactics.com/match/src/types"

func ErrorMessage(err error) *types.Message {
	return &types.Message{
		Action: types.ACTION_ERROR,
		Body: map[string]interface{}{
			"error": err.Error(),
		},
	}
}
