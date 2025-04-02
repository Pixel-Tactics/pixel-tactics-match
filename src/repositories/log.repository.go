package repositories

import (
	"errors"
	"strconv"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_LOG_PREFIX = "log_"
	LOG_COUNT       = BASE_LOG_PREFIX + "count_"
	LOG_DETAIL      = BASE_HERO_PREFIX + "detail_"
)

type SessionLogRepository interface {
	CountSessionLogs(tx databases.BadgerTx, sessionId string) (int, error)
	GetSessionLogs(tx databases.BadgerTx, sessionId string) ([]*models.SessionLog, error)
	AppendLog(tx databases.BadgerTx, sessionId string, obj *models.SessionLog) (*models.SessionLog, error)
}

type SessionLogRepositoryImpl struct {
	badger databases.Badger
}

func (repo *SessionLogRepositoryImpl) CountSessionLogs(tx databases.BadgerTx, sessionId string) (int, error) {
	query := databases.GetQuery(tx, repo.badger)

	var logCount int
	err := query.Get(LOG_COUNT+sessionId, &logCount)
	if err != nil {
		return 0, err
	}

	return logCount, nil
}

func (repo *SessionLogRepositoryImpl) GetSessionLogs(tx databases.BadgerTx, sessionId string) ([]*models.SessionLog, error) {
	query := databases.GetQuery(tx, repo.badger)

	var logCount int
	err := query.Get(LOG_COUNT+sessionId, &logCount)
	if err != nil {
		return nil, err
	}

	logs := make([]*models.SessionLog, logCount)
	for i := 0; i < logCount; i++ {
		var curLog *models.SessionLog
		err := query.Get(LOG_DETAIL+sessionId+"_"+strconv.Itoa(i), &curLog)
		if err != nil {
			return nil, err
		}
		logs = append(logs, curLog)
	}
	return logs, nil
}

func (repo *SessionLogRepositoryImpl) AppendLog(tx databases.BadgerTx, sessionId string, obj *models.SessionLog) (*models.SessionLog, error) {
	if tx == nil {
		return nil, errors.New("transaction object is null")
	}

	var logCount int
	err := tx.Get(LOG_COUNT+sessionId, &logCount)
	if err != nil {
		return nil, err
	}

	err = tx.Set(LOG_DETAIL+sessionId+"_"+strconv.Itoa(logCount), obj)
	if err != nil {
		return nil, err
	}

	err = tx.Set(LOG_COUNT+sessionId, logCount+1)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func NewSessionLogRepository(
	badger databases.Badger,
) SessionLogRepository {
	return &SessionLogRepositoryImpl{
		badger: badger,
	}
}
