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
	GetSessionById(tx databases.BadgerTx, sessionId string) (*models.Session, error)
	GetSessionByPlayerId(tx databases.BadgerTx, playerId string) (*models.Session, error)
	SaveSession(tx databases.BadgerTx, obj *models.Session) (*models.Session, error)
	GetInitialSession(tx databases.BadgerTx, sessionId string) (*models.Session, error)
	SaveInitialSession(tx databases.BadgerTx, obj *models.Session) (*models.Session, error)
	// CreateSession(tx databases.BadgerTx, params CreateSessionParams) (*models.Session, error)
	// UpdateSession(tx databases.BadgerTx, params UpdateSessionParams) (*models.Session, error)
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

// TODO: check whether not assigned returns error or nil.. if error, which error
func (repo *SessionRepositoryV2Impl) GetSessionById(tx databases.BadgerTx, sessionId string) (*models.Session, error) {
	query := databases.GetQuery(tx, repo.badger)

	var session models.Session
	err := query.Get(SESSION_ID_TO_SESSION+sessionId, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// TODO: check whether not assigned returns error or nil.. if error, which error
func (repo *SessionRepositoryV2Impl) GetSessionByPlayerId(tx databases.BadgerTx, playerId string) (*models.Session, error) {
	query := databases.GetQuery(tx, repo.badger)

	var sessionId string
	err := query.Get(PLAYERID_TO_SESSION+playerId, &sessionId)
	if err != nil {
		return nil, err
	}
	var session models.Session
	err = query.Get(SESSION_ID_TO_SESSION+sessionId, &session)
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

	err = tx.Set(PLAYERID_TO_SESSION+obj.PlayerIds[0], obj)
	if err != nil {
		return nil, err
	}

	err = tx.Set(PLAYERID_TO_SESSION+obj.PlayerIds[1], obj)
	if err != nil {
		return nil, err
	}
	return obj, err
}

func (repo *SessionRepositoryV2Impl) GetInitialSession(tx databases.BadgerTx, sessionId string) (*models.Session, error) {
	query := databases.GetQuery(tx, repo.badger)

	var session models.Session
	err := query.Get(SESSION_INITIAL+sessionId, &session)
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

// func (repo *SessionRepositoryV2Impl) CreateSession(tx databases.BadgerTx, params CreateSessionParams) (*models.Session, error) {
// 	if tx == nil {
// 		return nil, errors.New("transaction object is null")
// 	}

// 	sessionId := uuid.New().String()
// 	session := &models.Session{
// 		Id: sessionId,
// 		State: models.State{
// 			Id:   uuid.New().String(),
// 			Type: models.SessionStateMatchMaking,
// 		},
// 		AllowedHeroList: params.AvailableHeroList,
// 		PlayerIds: []string{
// 			params.PlayerId1,
// 			params.PlayerId2,
// 		},
// 	}
// 	err := tx.Set(SESSION_ID_TO_SESSION+sessionId, session)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = tx.Set(PLAYERID_TO_SESSION+session.PlayerIds[0], sessionId)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = tx.Set(PLAYERID_TO_SESSION+session.PlayerIds[1], sessionId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return session, nil
// }

// func (repo *SessionRepositoryV2Impl) UpdateSession(tx databases.BadgerTx, params UpdateSessionParams) (*models.Session, error) {
// 	query := databases.GetQuery(tx, repo.badger)

// 	session := repo.GetSessionById(nil, params.SessionId)
// 	if session == nil {
// 		return nil, errors.New("invalid session id")
// 	}

// 	session.State = params.State
// 	session.WinnerId = params.WinnerId

// 	err := query.Set(SESSION_ID_TO_SESSION+params.SessionId, session)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return session, nil
// }

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
