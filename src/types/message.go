package types

import (
	"encoding/json"
	"errors"
)

type MessageAction string

const (
	ACTION_GET  MessageAction = "GET"
	ACTION_POST MessageAction = "POST"
	ACTION_AUTH MessageAction = "AUTH"
)

type Message struct {
	Action MessageAction
	Body   map[string]string
}

func JsonBytesToMessage(jsonBytes []byte) (*Message, error) {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(jsonBytes, &raw)
	if err != nil {
		return nil, errors.New("json is invalid")
	}

	var action MessageAction
	err = json.Unmarshal(raw["action"], &action)
	if err != nil {
		return nil, errors.New("json is invalid")
	}

	var body map[string]string
	err = json.Unmarshal(raw["body"], &body)
	if err != nil {
		return nil, errors.New("json is invalid")
	}

	return &Message{Action: action, Body: body}, nil
}

func MessageToJsonBytes(message *Message) ([]byte, error) {
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return nil, errors.New("message is invalid")
	}

	return jsonBytes, nil
}
