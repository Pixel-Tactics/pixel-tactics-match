package integration_users

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

func GetUsernameFromToken(playerToken string) (string, error) {
	host := os.Getenv("USER_MICROSERVICE_URL")

	client := &http.Client{}
	req, err := http.NewRequest("GET", host+"/auth/current", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+playerToken)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New("invalid token")
	}

	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var body map[string]string
	json.Unmarshal(jsonBytes, &body)

	playerId, ok := body["username"]
	if !ok {
		return "", errors.New("invalid json body")
	}

	return playerId, nil
}
