package services

import (
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
)

type LogService interface {
	// Appends game log and emits `session_log_added` event
	AppendLog(tx databases.BadgerTx, sessionId string, obj *models.SessionLog) error
}

type LogServiceImpl struct {
	LogRepository repositories.SessionLogRepository
	EventManager  events.EventManager
}

func (service *LogServiceImpl) AppendLog(tx databases.BadgerTx, sessionId string, obj *models.SessionLog) error {
	_, err := service.LogRepository.AppendLog(tx, sessionId, obj)
	if err != nil {
		return err
	}
	return service.EventManager.Emit("session_log_added", obj)
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
