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
	GetPlayer(tx databases.BadgerTx, playerId string, sessionId string) *models.Player
	CreatePlayerForSession(tx databases.BadgerTx, sessionId string, playerId1 string, playerId2 string) error
	SetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string, heroList []heroes.BaseHeroEnum) error

	SetSessionService(sessionService SessionService)
}

type PlayerServiceImpl struct {
	PlayerRepository repositories.PlayerRepository

	SessionService SessionService
}

func (service *PlayerServiceImpl) GetPlayer(tx databases.BadgerTx, playerId string, sessionId string) *models.Player {
	return service.PlayerRepository.GetPlayerByIdAndSessionId(tx, playerId, sessionId)
}

func (service *PlayerServiceImpl) CreatePlayerForSession(tx databases.BadgerTx, sessionId string, playerId1 string, playerId2 string) error {
	session := service.SessionService.GetSessionById(tx, sessionId)
	if session == nil {
		return exceptions.SessionNotFound()
	}

	// TODO: check if player really exists as multilayer protection

	_, err := service.PlayerRepository.CreatePlayer(tx, repositories.CreatePlayerParams{
		Id:        playerId1,
		SessionId: sessionId,
	})
	if err != nil {
		return err
	}

	_, err = service.PlayerRepository.CreatePlayer(tx, repositories.CreatePlayerParams{
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
	player := service.GetPlayer(tx, playerId, sessionId)
	if player == nil {
		return errors.New("invalid player key")
	}

	_, err := service.PlayerRepository.UpdatePlayer(tx, repositories.UpdatePlayerParams{
		SessionId: sessionId,
		PlayerId:  playerId,
		HeroBases: heroList,
	})

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
