package actions

import (
	"encoding/json"
)

type ActionType = string

type Action interface {
	GetType() ActionType
}

func FromJson(jsonStr string, actionType ActionType) (Action, error) {
	jsonBytes := []byte(jsonStr)

	var err error
	var action Action
	switch actionType {
	case ActionTypeAttack:
		var attack AttackAction
		err = json.Unmarshal(jsonBytes, &attack)
		action = &attack
	case ActionTypeMove:
		var move MoveAction
		err = json.Unmarshal(jsonBytes, &move)
		action = &move
	}
	if err != nil {
		return nil, err
	}
	return action, nil
}
