package services

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
)

type PlayerService interface {
	// When no player found, nil player will be returned instead of error.
	GetPlayer(tx databases.BadgerTx, playerId string, sessionId string) (*models.Player, error)
	CreatePlayerForSession(tx databases.BadgerTx, sessionId string, playerId1 string, playerId2 string) error
	SetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string, heroList []heroes.BaseHeroEnum) error

	SetSessionService(sessionService SessionService)
}

type PlayerServiceImpl struct {
	PlayerRepository repositories.PlayerRepository

	SessionService SessionService
}

// When no player found, nil player will be returned instead of error.
func (service *PlayerServiceImpl) GetPlayer(tx databases.BadgerTx, playerId string, sessionId string) (*models.Player, error) {
	return service.PlayerRepository.GetPlayerByIdAndSessionId(tx, playerId, sessionId)
}

func (service *PlayerServiceImpl) CreatePlayerForSession(tx databases.BadgerTx, sessionId string, playerId1 string, playerId2 string) error {
	session, err := service.SessionService.GetSessionById(tx, sessionId)
	if err != nil {
		return err
	}
	if session == nil {
		return exceptions.SessionNotFound()
	}

	// TODO: check if player really exists as multilayer protection

	_, err = service.PlayerRepository.SavePlayer(tx, &models.Player{
		Id:        playerId1,
		SessionId: sessionId,
	})
	if err != nil {
		return err
	}

	_, err = service.PlayerRepository.SavePlayer(tx, &models.Player{
		Id:        playerId2,
		SessionId: sessionId,
	})
	if err != nil {
		service.PlayerRepository.DeletePlayer(tx, playerId1, sessionId)
		return err
	}

	return nil
}

func (service *PlayerServiceImpl) SetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string, heroList []heroes.BaseHeroEnum) error {
	player, err := service.GetPlayer(tx, playerId, sessionId)
	if err != nil {
		return err
	}
	if player == nil {
		return errors.New("invalid player key")
	}

	player.HeroBases = heroList
	_, err = service.PlayerRepository.SavePlayer(tx, player)
	return err
}

func (service *PlayerServiceImpl) SetSessionService(sessionService SessionService) {
	service.SessionService = sessionService
}

func NewPlayerService(
	playerRepository repositories.PlayerRepository,
	sessionService SessionService,
) PlayerService {
	return &PlayerServiceImpl{
		PlayerRepository: playerRepository,
		SessionService:   sessionService,
	}
}
