package repositories

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_SESSION_PREFIX   string = "session_"
	SESSION_ID_TO_SESSION string = BASE_SESSION_PREFIX + "id_"
	PLAYERID_TO_SESSION   string = BASE_SESSION_PREFIX + "username_"

	SESSION_INITIAL string = "initial_" + BASE_SESSION_PREFIX + "id_"
)

type SessionRepositoryV2 interface {
	// When no key found (or empty) for session, nil session will be returned instead of error.
	GetSessionById(tx databases.BadgerTx, sessionId string) (*models.Session, error)

	// When no key found (or empty) for session, nil session will be returned instead of error.
	GetSessionByPlayerId(tx databases.BadgerTx, playerId string) (*models.Session, error)

	SaveSession(tx databases.BadgerTx, obj *models.Session) (*models.Session, error)
	GetInitialSession(tx databases.BadgerTx, sessionId string) (*models.Session, error)
	SaveInitialSession(tx databases.BadgerTx, obj *models.Session) (*models.Session, error)
	DeleteSession(tx databases.BadgerTx, sessionId string) error
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

// When no key found (or empty) for session, nil session will be returned instead of error.
func (repo *SessionRepositoryV2Impl) GetSessionById(tx databases.BadgerTx, sessionId string) (*models.Session, error) {
	query := databases.GetQuery(tx, repo.badger)

	var session models.Session
	err := query.Get(SESSION_ID_TO_SESSION+sessionId, &session)
	if err != nil && err == databases.NotFoundException() {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// When no key found (or empty) for session, nil session will be returned instead of error.
func (repo *SessionRepositoryV2Impl) GetSessionByPlayerId(tx databases.BadgerTx, playerId string) (*models.Session, error) {
	query := databases.GetQuery(tx, repo.badger)

	var sessionId string
	err := query.Get(PLAYERID_TO_SESSION+playerId, &sessionId)
	if err != nil && err == databases.NotFoundException() {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var session models.Session
	err = query.Get(SESSION_ID_TO_SESSION+sessionId, &session)
	if err != nil && err == databases.NotFoundException() {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (repo *SessionRepositoryV2Impl) SaveSession(tx databases.BadgerTx, obj *models.Session) (*models.Session, error) {
	if tx == nil {
		return nil, errors.New("transaction object is null")
	}

	if len(obj.PlayerIds) != 2 {
		return nil, errors.New("invalid players")
	}

	err := tx.Set(SESSION_ID_TO_SESSION+obj.Id, obj)
	if err != nil {
		return nil, err
	}

	err = tx.Set(PLAYERID_TO_SESSION+obj.PlayerIds[0], obj.Id)
	if err != nil {
		return nil, err
	}

	err = tx.Set(PLAYERID_TO_SESSION+obj.PlayerIds[1], obj.Id)
	if err != nil {
		return nil, err
	}
	return obj, err
}

func (repo *SessionRepositoryV2Impl) GetInitialSession(tx databases.BadgerTx, sessionId string) (*models.Session, error) {
	query := databases.GetQuery(tx, repo.badger)

	var session models.Session
	err := query.Get(SESSION_INITIAL+sessionId, &session)
	if err != nil && err == databases.NotFoundException() {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (repo *SessionRepositoryV2Impl) SaveInitialSession(tx databases.BadgerTx, obj *models.Session) (*models.Session, error) {
	if tx == nil {
		return nil, errors.New("transaction object is null")
	}

	if len(obj.PlayerIds) != 2 {
		return nil, errors.New("invalid players")
	}

	err := tx.Set(SESSION_INITIAL+obj.Id, obj)
	if err != nil {
		return nil, err
	}
	return obj, err
}

func (repo *SessionRepositoryV2Impl) DeleteSession(tx databases.BadgerTx, sessionId string) error {
	if tx == nil {
		return errors.New("transaction object is null")
	}

	var session models.Session
	err := tx.Get(SESSION_ID_TO_SESSION+sessionId, &session)
	if err != nil {
		return err
	}

	err = tx.Delete(SESSION_ID_TO_SESSION + sessionId)
	if err != nil {
		return err
	}

	err = tx.Delete(PLAYERID_TO_SESSION + session.PlayerIds[0])
	if err != nil {
		return err
	}

	return tx.Delete(PLAYERID_TO_SESSION + session.PlayerIds[1])
}

func NewSessionRepositoryV2(
	badger databases.Badger,
) SessionRepositoryV2 {
	return &SessionRepositoryV2Impl{
		badger: badger,
	}
}
