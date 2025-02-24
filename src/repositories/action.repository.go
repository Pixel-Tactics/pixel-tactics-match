package repositories

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"pixeltactics.com/match/src/core/actions"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_ACTION_LOG_PREFIX = "hero_"
	ACTION_LIST            = BASE_ACTION_LOG_PREFIX + "list_"
)

type ActionLogRepository interface {
	GetSessionActionLogs(tx databases.BadgerTx, sessionId string) ([]*models.ActionLog, error)
	CreateActionLog(tx databases.BadgerTx, obj *models.ActionLog) (*models.ActionLog, error)
	UpdateActionLog(tx databases.BadgerTx, obj *models.ActionLog) (*models.ActionLog, error)
}

type ActionLogKey struct {
	SessionId string
	PlayerId  string
	Order     int
}

type SerializableActionLog struct {
	ActionLogKey
	Turn       int
	Type       actions.ActionType
	ActionJson string
}

type ActionLogRepositoryImpl struct {
	badger databases.Badger
}

func (repo *ActionLogRepositoryImpl) serializeActionLog(action *models.ActionLog) (*SerializableActionLog, error) {
	actionJson, err := json.Marshal(action.Action)
	if err != nil {
		return nil, err
	}
	return &SerializableActionLog{
		ActionLogKey: ActionLogKey{
			SessionId: action.SessionId,
			PlayerId:  action.PlayerId,
			Order:     action.Order,
		},
		Turn:       action.Turn,
		Type:       action.Action.GetType(),
		ActionJson: string(actionJson),
	}, nil
}

func (repo *ActionLogRepositoryImpl) GetSessionActionLogs(tx databases.BadgerTx, sessionId string) ([]*models.ActionLog, error) {
	query := databases.GetQuery(tx, repo.badger)

	var serializedLogs []*SerializableActionLog
	err := query.Get(ACTION_LIST+sessionId, &serializedLogs)
	if err != nil {
		return nil, err
	}
	log.Println(serializedLogs)
	logs := make([]*models.ActionLog, 0)

	for _, serializedLog := range serializedLogs {
		action, err := actions.FromJson(serializedLog.ActionJson, serializedLog.Type)
		if err != nil {
			return nil, err
		}
		logs = append(logs, &models.ActionLog{
			SessionId: serializedLog.SessionId,
			PlayerId:  serializedLog.PlayerId,
			Order:     serializedLog.Order,
			Action:    action,
		})
	}
	return logs, nil
}

func (repo *ActionLogRepositoryImpl) CreateActionLog(tx databases.BadgerTx, obj *models.ActionLog) (*models.ActionLog, error) {
	if tx == nil {
		return nil, errors.New("transaction object is null")
	}

	serializedObj, err := repo.serializeActionLog(obj)
	if err != nil {
		return nil, err
	}

	var serializedLogs []*SerializableActionLog
	err = tx.Get(ACTION_LIST+obj.SessionId, &serializedLogs)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return nil, err
		} else {
			serializedLogs = make([]*SerializableActionLog, 0)
		}
	}

	serializedLogs = append(serializedLogs, serializedObj)

	err = tx.Set(ACTION_LIST+obj.SessionId, serializedLogs)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (repo *ActionLogRepositoryImpl) UpdateActionLog(tx databases.BadgerTx, obj *models.ActionLog) (*models.ActionLog, error) {
	if tx == nil {
		return nil, errors.New("transaction object is null")
	}

	serializedObj, err := repo.serializeActionLog(obj)
	if err != nil {
		return nil, err
	}

	var serializedLogs []*SerializableActionLog
	err = tx.Get(ACTION_LIST+obj.SessionId, &serializedLogs)
	if err != nil {
		return nil, err
	}

	serializedLogs[obj.Order] = serializedObj

	err = tx.Set(ACTION_LIST+obj.SessionId, serializedLogs)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func NewActionLogRepository(
	badger databases.Badger,
) ActionLogRepository {
	return &ActionLogRepositoryImpl{
		badger: badger,
	}
}
