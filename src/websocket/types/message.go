package ws_types

import (
	"encoding/json"
	"errors"

	"pixeltactics.com/match/src/exceptions"
)

type MessageAction string

const (
	ACTION_CREATE_SESSION MessageAction = "CREATE_SESSION"
	ACTION_IS_IN_SESSION  MessageAction = "IS_IN_SESSION"
	ACTION_GET_SESSION    MessageAction = "GET_SESSION"
	ACTION_INVITE_SESSION MessageAction = "INVITE_SESSION"
	ACTION_START_SESSION  MessageAction = "START_SESSION"
	ACTION_PREPARE_PLAYER MessageAction = "PREPARE_PLAYER"
	ACTION_SERVER_TIME    MessageAction = "SERVER_TIME"
	ACTION_AUTH           MessageAction = "AUTH"
	ACTION_ERROR          MessageAction = "ERROR_FEEDBACK"
	ACTION_FEEDBACK       MessageAction = "FEEDBACK"
	ACTION_ENEMY_ACTION   MessageAction = "ENEMY_ACTION"
	ACTION_START_BATTLE   MessageAction = "START_BATTLE"
	ACTION_EXECUTE_ACTION MessageAction = "EXECUTE_ACTION"
	ACTION_APPLY_ACTION   MessageAction = "APPLY_ACTION"
	ACTION_END_TURN       MessageAction = "END_TURN"
)

type Message struct {
	Action     MessageAction          `json:"action"`
	Identifier string                 `json:"identifier"`
	Body       map[string]interface{} `json:"body"`
}

func JsonBytesToMessage(jsonBytes []byte) (*Message, error) {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(jsonBytes, &raw)
	if err != nil {
		return nil, exceptions.InvalidJsonError()
	}

	var action MessageAction
	err = json.Unmarshal(raw["action"], &action)
	if err != nil {
		return nil, exceptions.InvalidJsonError()
	}

	var identifier string
	err = json.Unmarshal(raw["identifier"], &identifier)
	if err != nil {
		return nil, exceptions.InvalidJsonError()
	}

	var body map[string]interface{}
	err = json.Unmarshal(raw["body"], &body)
	if err != nil {
		return nil, exceptions.InvalidJsonError()
	}

	return &Message{Action: action, Identifier: identifier, Body: body}, nil
}

func MessageToJsonBytes(message *Message) ([]byte, error) {
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return nil, errors.New("message is invalid")
	}

	return jsonBytes, nil
}
