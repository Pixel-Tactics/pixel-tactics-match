package repositories

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_PLAYER_PREFIX           = "player_"
	PLAYERID_SESSIONID_TO_PLAYER = BASE_PLAYER_PREFIX + "sessid_pid_"
)

type PlayerRepository interface {
	GetPlayerByIdAndSessionId(tx databases.BadgerTx, playerId string, sessionId string) *models.Player
	CreatePlayer(tx databases.BadgerTx, params CreatePlayerParams) (*models.Player, error)
	UpdatePlayer(tx databases.BadgerTx, params UpdatePlayerParams) (*models.Player, error)
	DeletePlayer(tx databases.BadgerTx, playerId string, sessionId string) error
}

type CreatePlayerParams struct {
	Id        string
	SessionId string
}

type UpdatePlayerParams struct {
	SessionId string
	PlayerId  string
	HeroBases []heroes.BaseHeroEnum
}

type PlayerRepositoryImpl struct {
	badger databases.Badger
}

func (repo *PlayerRepositoryImpl) GetPlayerByIdAndSessionId(tx databases.BadgerTx, playerId string, sessionId string) *models.Player {
	query := databases.GetQuery(tx, repo.badger)

	var player models.Player
	err := query.Get(PLAYERID_SESSIONID_TO_PLAYER+sessionId+"_"+playerId, &player)
	if err != nil {
		return nil
	}
	return &player
}

func (repo *PlayerRepositoryImpl) CreatePlayer(tx databases.BadgerTx, params CreatePlayerParams) (*models.Player, error) {
	query := databases.GetQuery(tx, repo.badger)

	player := &models.Player{
		Id:        params.Id,
		SessionId: params.SessionId,
	}
	err := query.Set(PLAYERID_SESSIONID_TO_PLAYER+params.SessionId+"_"+params.Id, player)
	if err != nil {
		return nil, err
	}
	return player, nil
}

func (repo *PlayerRepositoryImpl) UpdatePlayer(tx databases.BadgerTx, params UpdatePlayerParams) (*models.Player, error) {
	player := repo.GetPlayerByIdAndSessionId(tx, params.PlayerId, params.SessionId)
	if player == nil {
		return nil, errors.New("invalid player key")
	}

	player.HeroBases = params.HeroBases

	err := tx.Set(PLAYERID_SESSIONID_TO_PLAYER+params.SessionId+"_"+params.PlayerId, player)
	if err != nil {
		return nil, err
	}
	return player, nil
}

func (repo *PlayerRepositoryImpl) DeletePlayer(tx databases.BadgerTx, playerId string, sessionId string) error {
	var player models.Player
	err := tx.Get(PLAYERID_SESSIONID_TO_PLAYER+sessionId+"_"+playerId, &player)
	if err != nil {
		return err
	}
	return tx.Delete(PLAYERID_SESSIONID_TO_PLAYER + sessionId + "_" + playerId)
}

func NewPlayerRepositoryImpl(
	badger databases.Badger,
) *PlayerRepositoryImpl {
	return &PlayerRepositoryImpl{
		badger: badger,
	}
}
