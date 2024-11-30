package services

import (
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
)

type MapService interface {
	GetSessionMap(sessionId string) (*models.Map, error)
	GenerateMap(sessionId string) (*models.Map, error)
}

type MapServiceImpl struct {
	mapRepository repositories.MapRepository
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
