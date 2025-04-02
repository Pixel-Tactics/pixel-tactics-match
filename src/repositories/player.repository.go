package repositories

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_PLAYER_PREFIX           = "player_"
	PLAYERID_SESSIONID_TO_PLAYER = BASE_PLAYER_PREFIX + "sessid_pid_"
)

type PlayerRepository interface {
	// When no key found (or empty) for player, nil player will be returned instead of error.
	GetPlayerByIdAndSessionId(tx databases.BadgerTx, playerId string, sessionId string) (*models.Player, error)
	SavePlayer(tx databases.BadgerTx, obj *models.Player) (*models.Player, error)
	DeletePlayer(tx databases.BadgerTx, playerId string, sessionId string) error
}

type PlayerRepositoryImpl struct {
	badger databases.Badger
}

// When no key found (or empty) for player, nil player will be returned instead of error.
func (repo *PlayerRepositoryImpl) GetPlayerByIdAndSessionId(tx databases.BadgerTx, playerId string, sessionId string) (*models.Player, error) {
	query := databases.GetQuery(tx, repo.badger)

	var player models.Player
	err := query.Get(PLAYERID_SESSIONID_TO_PLAYER+sessionId+"_"+playerId, &player)
	if err != nil && err == databases.NotFoundException() {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (repo *PlayerRepositoryImpl) SavePlayer(tx databases.BadgerTx, obj *models.Player) (*models.Player, error) {
	query := databases.GetQuery(tx, repo.badger)

	err := query.Set(PLAYERID_SESSIONID_TO_PLAYER+obj.SessionId+"_"+obj.Id, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (repo *PlayerRepositoryImpl) DeletePlayer(tx databases.BadgerTx, playerId string, sessionId string) error {
	if tx == nil {
		return errors.New("transaction object is null")
	}

	var player models.Player
	err := tx.Get(PLAYERID_SESSIONID_TO_PLAYER+sessionId+"_"+playerId, &player)
	if err != nil {
		return err
	}
	return tx.Delete(PLAYERID_SESSIONID_TO_PLAYER + sessionId + "_" + playerId)
}

func NewPlayerRepository(
	badger databases.Badger,
) PlayerRepository {
	return &PlayerRepositoryImpl{
		badger: badger,
	}
}
