package repositories

import (
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_MAP_PREFIX  = "map_"
	SESSIONID_TO_MAP = BASE_PLAYER_PREFIX + "session_"
)

type MapRepository interface {
	GetMapBySessionId(sessionId string) *models.Map
	CreateMap(params CreateMapParams) (*models.Map, error)
	DeleteMap(sessionId string) error
}

type CreateMapParams struct {
	SessionId string
	Structure [][]int
}

type MapRepositoryImpl struct {
	badger databases.Badger
}

func (repo *MapRepositoryImpl) GetMapBySessionId(sessionId string) *models.Map {
	var sessionMap models.Map
	err := repo.badger.Get(SESSIONID_TO_MAP+sessionId, &sessionMap)
	if err != nil {
		return nil
	}
	return &sessionMap
}

func (repo *MapRepositoryImpl) CreateMap(params CreateMapParams) (*models.Map, error) {
	sessionMap := &models.Map{
		SessionId: params.SessionId,
		Structure: params.Structure,
	}
	err := repo.badger.Set(SESSIONID_TO_MAP+params.SessionId, sessionMap)
	if err != nil {
		return nil, err
	}
	return sessionMap, nil
}

func (repo *MapRepositoryImpl) DeleteMap(sessionId string) error {
	var sessionMap models.Map
	err := repo.badger.Get(SESSIONID_TO_MAP+sessionId, &sessionMap)
	if err != nil {
		return err
	}
	return repo.badger.Delete(SESSIONID_TO_MAP + sessionId)
}

func NewMapRepositoryImpl(
	badger databases.Badger,
) *MapRepositoryImpl {
	return &MapRepositoryImpl{
		badger: badger,
	}
}
