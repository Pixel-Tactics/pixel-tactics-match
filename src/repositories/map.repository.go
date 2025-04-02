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
	// When no key found (or empty) for map, nil map will be returned instead of error.
	GetMapBySessionId(tx databases.BadgerTx, sessionId string) (*models.Map, error)
	SaveMap(tx databases.BadgerTx, obj *models.Map) (*models.Map, error)
	DeleteMap(tx databases.BadgerTx, sessionId string) error
}

type CreateMapParams struct {
	SessionId string
	Structure [][]int
}

type MapRepositoryImpl struct {
	badger databases.Badger
}

// When no key found (or empty) for map, nil map will be returned instead of error.
func (repo *MapRepositoryImpl) GetMapBySessionId(tx databases.BadgerTx, sessionId string) (*models.Map, error) {
	query := databases.GetQuery(tx, repo.badger)

	var sessionMap models.Map
	err := query.Get(SESSIONID_TO_MAP+sessionId, &sessionMap)
	if err != nil && err == databases.NotFoundException() {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sessionMap, nil
}

func (repo *MapRepositoryImpl) SaveMap(tx databases.BadgerTx, obj *models.Map) (*models.Map, error) {
	query := databases.GetQuery(tx, repo.badger)

	err := query.Set(SESSIONID_TO_MAP+obj.SessionId, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
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

func NewMapRepository(
	badger databases.Badger,
) MapRepository {
	return &MapRepositoryImpl{
		badger: badger,
	}
}
