package services

import (
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
	"pixeltactics.com/match/src/utils/physics"
)

func getMapTemplate() [][]int {
	return [][]int{
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
	}
}

type MapService interface {
	// SessionNotFound error will be returned when map doesn't exist.
	GetSessionMap(tx databases.BadgerTx, sessionId string) (*models.Map, error)

	GenerateMap(tx databases.BadgerTx, sessionId string) (*models.Map, error)
	GetInitialState(tx databases.BadgerTx, sessionId string) (*models.Map, error)
	IsPointOpen(tx databases.BadgerTx, session *models.Session, pos physics.Point) (bool, error)

	SetHeroService(heroService HeroService)
	SetMap(tx databases.BadgerTx, newMap *models.Map) error
}

type MapServiceImpl struct {
	MapRepository repositories.MapRepository
	HeroService   HeroService
}

// SessionNotFound error will be returned when map doesn't exist.
func (service *MapServiceImpl) GetSessionMap(tx databases.BadgerTx, sessionId string) (*models.Map, error) {
	sessionMap, err := service.MapRepository.GetMapBySessionId(tx, sessionId)
	if err != nil {
		return nil, err
	}
	if sessionMap == nil {
		return nil, exceptions.SessionNotFound()
	}
	return sessionMap, nil
}

func (service *MapServiceImpl) GenerateMap(tx databases.BadgerTx, sessionId string) (*models.Map, error) {
	return service.MapRepository.SaveMap(tx, &models.Map{
		SessionId: sessionId,
		Structure: getMapTemplate(),
	})
}

func (service *MapServiceImpl) IsPointOpen(tx databases.BadgerTx, session *models.Session, pos physics.Point) (bool, error) {
	playerId1 := session.PlayerIds[0]
	playerId2 := session.PlayerIds[1]

	heroes1, err := service.HeroService.GetPlayerHeroes(tx, session.Id, playerId1)
	if err != nil {
		return false, err
	}

	heroes2, err := service.HeroService.GetPlayerHeroes(tx, session.Id, playerId2)
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
	sessionMap, err := service.GetSessionMap(tx, session.Id)
	if err != nil {
		return false, err
	}
	curValue := sessionMap.Structure[pos.Y][pos.X]
	return curValue != 2, nil
}

func (service *MapServiceImpl) GetInitialState(tx databases.BadgerTx, sessionId string) (*models.Map, error) {
	return &models.Map{
		SessionId: sessionId,
		Structure: getMapTemplate(),
	}, nil
}

func (service *MapServiceImpl) SetMap(tx databases.BadgerTx, newMap *models.Map) error {
	_, err := service.MapRepository.SaveMap(tx, newMap)
	return err
}

func (service *MapServiceImpl) SetHeroService(heroService HeroService) {
	service.HeroService = heroService
}

func NewMapService(
	mapRepository repositories.MapRepository,
	heroService HeroService,
) MapService {
	return &MapServiceImpl{
		MapRepository: mapRepository,
		HeroService:   heroService,
	}
}
