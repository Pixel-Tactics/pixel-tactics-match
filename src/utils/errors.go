package utils

import "pixeltactics.com/match/src/types"

func ErrorMessage(identifier string, err error) *types.Message {
	return &types.Message{
		Action: types.ACTION_ERROR,
		Body: map[string]interface{}{
			"error": err.Error(),
		},
	}
}
