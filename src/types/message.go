package types

import (
	"encoding/json"
	"errors"

	"pixeltactics.com/match/src/exceptions"
)

type MessageAction string

const (
	ACTION_CREATE_SESSION MessageAction = "CREATE_SESSION"
	ACTION_GET_SESSION    MessageAction = "GET_SESSION"
	ACTION_INVITE_SESSION MessageAction = "INVITE_SESSION"
	ACTION_START_SESSION  MessageAction = "START_SESSION"
	ACTION_AUTH           MessageAction = "AUTH"
	ACTION_ERROR          MessageAction = "ERROR"
	ACTION_FEEDBACK       MessageAction = "FEEDBACK"
)

type Message struct {
	Action     MessageAction
	Identifier string
	Body       map[string]interface{}
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
