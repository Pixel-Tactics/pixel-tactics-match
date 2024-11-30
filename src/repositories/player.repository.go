package repositories

import (
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_PLAYER_PREFIX           = "player_"
	PLAYERID_SESSIONID_TO_PLAYER = BASE_PLAYER_PREFIX + "sessid_pid_"
)

type PlayerRepository interface{}

type CreatePlayerParams struct {
	Id        string
	SessionId string
}

type PlayerRepositoryImpl struct {
	badger databases.Badger
}

func (repo *PlayerRepositoryImpl) GetPlayerByIdAndSessionId(playerId string, sessionId string) *models.Player {
	var player models.Player
	err := repo.badger.Get(PLAYERID_SESSIONID_TO_PLAYER+sessionId+"_"+playerId, &player)
	if err != nil {
		return nil
	}
	return &player
}

func (repo *PlayerRepositoryImpl) CreatePlayer(params CreatePlayerParams) (*models.Player, error) {
	player := &models.Player{
		Id:        params.Id,
		SessionId: params.SessionId,
	}
	err := repo.badger.Set(PLAYERID_SESSIONID_TO_PLAYER+params.SessionId+"_"+params.Id, player)
	if err != nil {
		return nil, err
	}
	return player, nil
}

func (repo *PlayerRepositoryImpl) DeletePlayer(playerId string, sessionId string) error {
	var player models.Player
	err := repo.badger.Get(PLAYERID_SESSIONID_TO_PLAYER+sessionId+"_"+playerId, &player)
	if err != nil {
		return err
	}
	return repo.badger.Delete(PLAYERID_SESSIONID_TO_PLAYER + sessionId + "_" + playerId)
}

func NewPlayerRepositoryImpl(
	badger databases.Badger,
) *PlayerRepositoryImpl {
	return &PlayerRepositoryImpl{
		badger: badger,
	}
}
