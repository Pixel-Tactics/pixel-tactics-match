package repositories

import (
	"errors"

	"github.com/google/uuid"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_SESSION_PREFIX   string = "session_"
	SESSION_ID_TO_SESSION string = BASE_SESSION_PREFIX + "id_"
	PLAYERID_TO_SESSION   string = BASE_SESSION_PREFIX + "username_"
)

type SessionRepositoryV2 interface {
	GetSessionById(sessionId string) *models.Session
	GetSessionByPlayerId(playerId string) *models.Session
	CreateSession(params CreateSessionParams) (*models.Session, error)
	UpdateSession(params UpdateSessionParams) (*models.Session, error)
	DeleteSession(sessionId string) error
}

type CreateSessionParams struct {
	PlayerId1         string
	PlayerId2         string
	AvailableHeroList []string
	State             models.State
}

type UpdateSessionParams struct {
	SessionId string
	State     models.State
	WinnerId  *string
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
		Id: sessionId,
		State: models.State{
			Id:   uuid.New().String(),
			Type: models.SessionStateMatchMaking,
		},
		AllowedHeroList: params.AvailableHeroList,
		PlayerIds: []string{
			params.PlayerId1,
			params.PlayerId2,
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

func (repo *SessionRepositoryV2Impl) UpdateSession(params UpdateSessionParams) (*models.Session, error) {
	session := repo.GetSessionById(params.SessionId)
	if session == nil {
		return nil, errors.New("invalid session id")
	}

	session.State = params.State
	session.WinnerId = params.WinnerId

	err := repo.badger.Set(SESSION_ID_TO_SESSION+params.SessionId, session)
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
