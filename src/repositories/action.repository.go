package repositories

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"pixeltactics.com/match/src/core/actions"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_ACTION_LOG_PREFIX = "hero_"
	ACTION_LIST            = BASE_ACTION_LOG_PREFIX + "list_"
	ACTION_DETAIL          = BASE_ACTION_LOG_PREFIX + "detail_"
)

type ActionLogRepository interface {
	GetSessionActionLogs(sessionId string) ([]*models.ActionLog, error)
	CreateActionLog(obj *models.ActionLog) (*models.ActionLog, error)
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

func (repo *ActionLogRepositoryImpl) GetSessionActionLogs(sessionId string) ([]*models.ActionLog, error) {
	var serializedLogs []*SerializableActionLog
	err := repo.badger.Get(ACTION_LIST+sessionId, &serializedLogs)
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

func (repo *ActionLogRepositoryImpl) CreateActionLog(obj *models.ActionLog) (*models.ActionLog, error) {
	serializedObj, err := repo.serializeActionLog(obj)
	if err != nil {
		return nil, err
	}

	tx := repo.badger.NewReadWriteTransaction()
	defer tx.Discard()

	detailKey := ACTION_DETAIL + "_" + serializedObj.SessionId + "_" + serializedObj.PlayerId + "_" + strconv.Itoa(serializedObj.Order)

	var existsCheck *SerializableActionLog
	err = tx.Get(detailKey, existsCheck)
	if err == nil {
		return nil, errors.New("key is used")
	}

	err = tx.Set(detailKey, serializedObj)
	if err != nil {
		return nil, err
	}

	err = repo.addActionLogIntoList(tx, serializedObj)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (repo *ActionLogRepositoryImpl) addActionLogIntoList(tx databases.BadgerTx, obj *SerializableActionLog) error {
	listKey := ACTION_LIST + obj.SessionId

	var logs []*SerializableActionLog
	err := tx.Get(listKey, &logs)
	if err != nil {
		logs = make([]*SerializableActionLog, 0)
	}
	logs = append(logs, obj)
	return tx.Set(listKey, logs)
}

func NewActionLogRepositoryImpl(
	badger databases.Badger,
) *ActionLogRepositoryImpl {
	return &ActionLogRepositoryImpl{
		badger: badger,
	}
}
