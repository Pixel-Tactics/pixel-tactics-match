package utils

import (
	"encoding/json"

	"pixeltactics.com/match/src/exceptions"
)

func MapToObject(data map[string]interface{}, obj any) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return exceptions.InvalidDataError()
	}

	err = json.Unmarshal(encoded, obj)
	if err != nil {
		return exceptions.InvalidDataError()
	}

	return nil
}

func ObjectToMap(obj any) (map[string]interface{}, error) {
	encoded, err := json.Marshal(obj)
	if err != nil {
		return nil, exceptions.InvalidDataError()
	}

	var data map[string]interface{}
	err = json.Unmarshal(encoded, &data)
	if err != nil {
		return nil, exceptions.InvalidDataError()
	}

	return data, nil
}
