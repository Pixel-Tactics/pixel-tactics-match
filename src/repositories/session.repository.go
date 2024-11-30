package repositories

import (
	"github.com/google/uuid"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_SESSION_PREFIX   string = "session_"
	SESSION_ID_TO_SESSION string = BASE_SESSION_PREFIX + "id_"
	PLAYERID_TO_SESSION   string = BASE_SESSION_PREFIX + "username_"
)

type SessionRepositoryV2 interface{}

type CreateSessionParams struct {
	Username1 string
	Username2 string
	HeroList  []string
}

type SessionRepositoryV2Impl struct {
	badger databases.Badger
}

func (repo *SessionRepositoryV2Impl) GetSessionById(sessionId string) *models.Session {
	var session models.Session
	err := repo.badger.Get(SESSION_ID_TO_SESSION+sessionId, &session)
	if err != nil {
		return nil
	}
	return &session
}

func (repo *SessionRepositoryV2Impl) GetSessionByPlayerId(playerId string) *models.Session {
	var sessionId string
	err := repo.badger.Get(PLAYERID_TO_SESSION+playerId, &sessionId)
	if err != nil {
		return nil
	}
	var session models.Session
	err = repo.badger.Get(SESSION_ID_TO_SESSION+sessionId, &session)
	if err != nil {
		return nil
	}
	return &session
}

func (repo *SessionRepositoryV2Impl) CreateSession(params CreateSessionParams) (*models.Session, error) {
	sessionId := uuid.New().String()
	session := &models.Session{
		Id:       sessionId,
		State:    models.SessionStateMatchMaking,
		HeroList: params.HeroList,
		PlayerIds: []string{
			params.Username1,
			params.Username2,
		},
	}
	err := repo.badger.BatchSet([]*databases.BadgerSetParams{
		{
			Key:   SESSION_ID_TO_SESSION + sessionId,
			Value: session,
		},
		{
			Key:   PLAYERID_TO_SESSION + session.PlayerIds[0],
			Value: sessionId,
		},
		{
			Key:   PLAYERID_TO_SESSION + session.PlayerIds[1],
			Value: sessionId,
		},
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (repo *SessionRepositoryV2Impl) DeleteSession(sessionId string) error {
	var session models.Session
	err := repo.badger.Get(SESSION_ID_TO_SESSION+sessionId, &session)
	if err != nil {
		return err
	}
	return repo.badger.BatchDelete([]*databases.BadgerDeleteParams{
		{
			Key: SESSION_ID_TO_SESSION + sessionId,
		},
		{
			Key: PLAYERID_TO_SESSION + session.PlayerIds[0],
		},
		{
			Key: PLAYERID_TO_SESSION + session.PlayerIds[1],
		},
	})
}

func NewSessionRepositoryV2Impl(
	badger databases.Badger,
) *SessionRepositoryV2Impl {
	return &SessionRepositoryV2Impl{
		badger: badger,
	}
}
