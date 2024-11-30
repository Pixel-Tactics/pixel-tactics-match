package repositories

import (
	"github.com/google/uuid"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_PREFIX           string = "session_"
	SESSION_ID_TO_SESSION string = BASE_PREFIX + "id_"
	USERNAME_TO_SESSION   string = BASE_PREFIX + "username_"
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

func (repo *SessionRepositoryV2Impl) GetSessionByUsername(username string) *models.Session {
	var sessionId string
	err := repo.badger.Get(USERNAME_TO_SESSION+username, &sessionId)
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
		PlayerUsername: []string{
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
			Key:   USERNAME_TO_SESSION + session.PlayerUsername[0],
			Value: sessionId,
		},
		{
			Key:   USERNAME_TO_SESSION + session.PlayerUsername[1],
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
			Key: USERNAME_TO_SESSION + session.PlayerUsername[0],
		},
		{
			Key: USERNAME_TO_SESSION + session.PlayerUsername[1],
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
