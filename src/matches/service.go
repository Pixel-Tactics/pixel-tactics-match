package matches

import (
	"errors"
)

var repo = GetSessionRepository()

type MatchService struct{}

type CreateSessionRequestDTO struct {
	PlayerId   string `json:"playerId"`
	OpponentId string `json:"opponentId"`
}

func (service MatchService) CreateSession(data CreateSessionRequestDTO) (*Session, error) {
	// TODO: auth middleware (for player)
	// TODO: check by fetch from account service (to check opponent if they exists)
	player := &Player{
		Id: data.PlayerId,
	}
	opponent := &Player{
		Id: data.OpponentId,
	}

	session, err := repo.CreateSession(player, opponent)
	if err != nil {
		return nil, err
	}

	return session, nil
}

type GetSessionRequestDTO struct {
	sessionId     string
	sessionSecret string
}

type GetSessionResponseDTO struct {
	MatchMap          MatchMap `json:"map"`
	AvailableHeroList []string `json:"available"`
}

func (service MatchService) GetSession(data GetSessionRequestDTO) (*GetSessionResponseDTO, error) {
	session := repo.GetSessionById(data.sessionId)
	if session == nil {
		return nil, errors.New("session not found")
	}

	if session.Secret != data.sessionSecret {
		return nil, errors.New("session secret is not valid")
	}

	var res GetSessionResponseDTO
	res.MatchMap = *session.MatchMap
	res.AvailableHeroList = *session.AvailableHeroList

	return &res, nil
}
