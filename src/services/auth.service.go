package services

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"pixeltactics.com/match/src/config"
)

type AuthService interface {
	GetUsernameFromToken(playerToken string) (string, error)
}

type AuthServiceImpl struct{}

func (service *AuthServiceImpl) GetUsernameFromToken(playerToken string) (string, error) {
	client := &http.Client{}
	host := config.UserServiceUrl
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

func NewAuthService() AuthService {
	return &AuthServiceImpl{}
}
