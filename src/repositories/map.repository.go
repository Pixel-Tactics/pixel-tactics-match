package repositories

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_MAP_PREFIX  = "map_"
	SESSIONID_TO_MAP = BASE_PLAYER_PREFIX + "session_"
)

type MapRepository interface {
	GetMapBySessionId(tx databases.BadgerTx, sessionId string) *models.Map
	CreateMap(tx databases.BadgerTx, params CreateMapParams) (*models.Map, error)
	DeleteMap(tx databases.BadgerTx, sessionId string) error
}

type CreateMapParams struct {
	SessionId string
	Structure [][]int
}

type MapRepositoryImpl struct {
	badger databases.Badger
}

func (repo *MapRepositoryImpl) GetMapBySessionId(tx databases.BadgerTx, sessionId string) *models.Map {
	query := databases.GetQuery(tx, repo.badger)

	var sessionMap models.Map
	err := query.Get(SESSIONID_TO_MAP+sessionId, &sessionMap)
	if err != nil {
		return nil
	}
	return &sessionMap
}

func (repo *MapRepositoryImpl) CreateMap(tx databases.BadgerTx, params CreateMapParams) (*models.Map, error) {
	query := databases.GetQuery(tx, repo.badger)

	sessionMap := &models.Map{
		SessionId: params.SessionId,
		Structure: params.Structure,
	}
	err := query.Set(SESSIONID_TO_MAP+params.SessionId, sessionMap)
	if err != nil {
		return nil, err
	}
	return sessionMap, nil
}

func (repo *MapRepositoryImpl) DeleteMap(tx databases.BadgerTx, sessionId string) error {
	if tx == nil {
		return errors.New("transaction object is null")
	}

	var sessionMap models.Map
	err := tx.Get(SESSIONID_TO_MAP+sessionId, &sessionMap)
	if err != nil {
		return err
	}
	return tx.Delete(SESSIONID_TO_MAP + sessionId)
}

func NewMapRepositoryImpl(
	badger databases.Badger,
) *MapRepositoryImpl {
	return &MapRepositoryImpl{
		badger: badger,
	}
}
