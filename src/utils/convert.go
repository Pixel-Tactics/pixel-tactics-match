package utils

import (
	"encoding/json"
	"errors"
)

func MapToObject(data map[string]string, obj any) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return errors.New("data is invalid")
	}

	err = json.Unmarshal(encoded, obj)
	if err != nil {
		return errors.New("data is invalid")
	}

	return nil
}
