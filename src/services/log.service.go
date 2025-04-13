package services

import (
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
)

const LOG_ADDED_EVENT = "session_log_added"

type LogService interface {
	// If session id doesn't have log, then empty log will be returned.
	GetSessionLogs(tx databases.BadgerTx, sessionId string) ([]*models.SessionLog, error)

	// Appends game log and emits `session_log_added` event
	AppendLog(tx databases.BadgerTx, session *models.Session, obj *models.SessionLog) error
}

type LogAddedEvent struct {
	SessionId string
	PlayerIDs []string
	Log       *models.SessionLog
	LogCount  int
}

type LogServiceImpl struct {
	LogRepository repositories.SessionLogRepository
	EventManager  events.EventManager
}

func (service *LogServiceImpl) GetSessionLogs(tx databases.BadgerTx, sessionId string) ([]*models.SessionLog, error) {
	return service.LogRepository.GetSessionLogs(tx, sessionId)
}

func (service *LogServiceImpl) AppendLog(tx databases.BadgerTx, session *models.Session, obj *models.SessionLog) error {
	// TODO: must handle failures like failed to send to broker, etc.
	_, logCount, err := service.LogRepository.AppendLog(tx, obj.SessionId, obj)
	if err != nil {
		return err
	}
	if logCount < 3 && session.State.Type != models.SessionStateEnded {
		return nil
	}

	event := &LogAddedEvent{
		SessionId: session.Id,
		PlayerIDs: session.PlayerIds,
		Log:       obj,
		LogCount:  logCount,
	}
	return service.EventManager.Emit(LOG_ADDED_EVENT, event)
}

func NewLogService(
	logRepository repositories.SessionLogRepository,
	eventManager events.EventManager,
) LogService {
	return &LogServiceImpl{
		LogRepository: logRepository,
		EventManager:  eventManager,
	}
}
