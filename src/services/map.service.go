package services

import (
	"errors"

	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
	"pixeltactics.com/match/src/utils/physics"
)

type MapService interface {
	GetSessionMap(sessionId string) (*models.Map, error)
	GenerateMap(sessionId string) (*models.Map, error)
	IsPointOpen(sessionId string, pos physics.Point) (bool, error)
}

type MapServiceImpl struct {
	mapRepository  repositories.MapRepository
	sessionService SessionService
	heroService    HeroService
}

func (service *MapServiceImpl) GetSessionMap(sessionId string) (*models.Map, error) {
	sessionMap := service.mapRepository.GetMapBySessionId(sessionId)
	if sessionMap == nil {
		return nil, exceptions.SessionNotFound()
	}
	return sessionMap, nil
}

func (service *MapServiceImpl) GenerateMap(sessionId string) (*models.Map, error) {
	return service.mapRepository.CreateMap(repositories.CreateMapParams{
		SessionId: sessionId,
		Structure: [][]int{
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 2, 2, 2, 2, 2, 2},
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 4, 3, 1, 1, 1, 1},
			{2, 2, 2, 2, 2, 2, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
			{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		},
	})
}

func (service *MapServiceImpl) IsPointOpen(sessionId string, pos physics.Point) (bool, error) {
	session := service.sessionService.GetSessionById(sessionId)
	if session == nil {
		return false, errors.New("invalid session id")
	}
	playerId1 := session.PlayerIds[0]
	playerId2 := session.PlayerIds[1]

	heroes1, err := service.heroService.GetPlayerHeroes(session.Id, playerId1)
	if err != nil {
		return false, err
	}

	heroes2, err := service.heroService.GetPlayerHeroes(session.Id, playerId2)
	if err != nil {
		return false, err
	}

	for _, hero := range heroes1 {
		if hero.Position.Equals(pos) {
			return false, nil
		}
	}
	for _, hero := range heroes2 {
		if hero.Position.Equals(pos) {
			return false, nil
		}
	}
	sessionMap, err := service.GetSessionMap(sessionId)
	if err != nil {
		return false, err
	}
	curValue := sessionMap.Structure[pos.Y][pos.X]
	return curValue != 2, nil
}
