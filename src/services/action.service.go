package services

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
	"pixeltactics.com/match/src/utils/algorithms"
	"pixeltactics.com/match/src/utils/physics"
)

type ActionService interface {
	Move(sessionId string, playerId string, baseHero string, directions []physics.Direction) error
	Attack(sessionId string, playerId string, srcBaseHero string, dstBaseHero string) error
}

type ActionServiceImpl struct {
	mapService     MapService
	heroService    HeroService
	sessionService SessionService

	sessionLogRepository repositories.SessionLogRepository
	transactionManager   databases.TransactionManager
	eventManager         events.EventManager
}

func (service *ActionServiceImpl) Move(sessionId string, playerId string, baseHero string, directions []physics.Direction) error {
	tx := service.transactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	session := service.sessionService.GetSessionById(tx, sessionId)
	if session == nil {
		return errors.New("invalid session id")
	}

	activePlayerId, err := session.GetActivePlayer()
	if err != nil {
		return err
	}

	srcHero := service.heroService.GetPlayerHero(tx, sessionId, playerId, baseHero)
	if srcHero == nil {
		return errors.New("invalid hero")
	}

	if !srcHero.CanMoveOnTurn(activePlayerId, session.CurrentTurn) {
		return errors.New("hero cannot move")
	}

	srcHeroStats := service.heroService.GetStats(srcHero.BaseHero)
	if len(directions) > srcHeroStats.BaseStats.MoveRange {
		return errors.New("invalid movement range")
	}

	curPos := srcHero.Position
	for _, dir := range directions {
		dirPoint := physics.GetPointFromDirection(dir)
		curPos = curPos.Add(dirPoint)
		isPointOpen, err := service.mapService.IsPointOpen(tx, session, curPos)
		if err != nil {
			return err
		}
		if !isPointOpen {
			return errors.New("point is occupied")
		}
	}

	_, err = service.sessionLogRepository.AppendLog(tx, sessionId, &models.SessionLog{
		Type: "MOVE",
		Data: map[string]interface{}{
			"playerId": playerId,
			"baseHero": baseHero,
			"point":    curPos,
		},
	})
	if err != nil {
		return err
	}

	err = service.eventManager.Emit("MOVE", &MoveEvent{
		Tx:          tx,
		SrcHero:     srcHero,
		Position:    curPos,
		CurrentTurn: session.CurrentTurn,
	})
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil

	// service.heroService.MoveHero(tx, actionLog.Order, srcHero, curPos)

	// err = tx.Commit()
	// if err != nil {
	// 	return err
	// }
	// return nil
}

func (service *ActionServiceImpl) Attack(sessionId string, playerId string, srcBaseHero string, dstBaseHero string) error {
	tx := service.transactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	session := service.sessionService.GetSessionById(tx, sessionId)
	if session == nil {
		return errors.New("invalid session id")
	}

	srcHero := service.heroService.GetPlayerHero(tx, sessionId, playerId, srcBaseHero)
	if srcHero == nil {
		return errors.New("invalid source hero")
	}

	otherPlayerId, _ := session.GetOtherPlayerId(playerId)
	dstHero := service.heroService.GetPlayerHero(tx, sessionId, otherPlayerId, dstBaseHero)
	if dstHero == nil {
		return errors.New("invalid destination hero")
	}

	activePlayerId, err := session.GetActivePlayer()
	if err != nil {
		return err
	}

	if !srcHero.CanAttackOnTurn(activePlayerId, session.CurrentTurn) {
		return errors.New("hero cannot attack")
	}

	srcHeroStats := service.heroService.GetStats(srcHero.BaseHero)

	attackRange := srcHeroStats.BaseStats.AttackRange
	damage := srcHeroStats.BaseStats.Damage

	sessionMap, err := service.mapService.GetSessionMap(tx, sessionId)
	if err != nil {
		return err
	}

	dist, err := algorithms.CheckDistance(sessionMap.Structure, srcHero.Position, dstHero.Position)
	if dist > attackRange || err != nil {
		return errors.New("target out of range")
	}

	_, err = service.sessionLogRepository.AppendLog(tx, sessionId, &models.SessionLog{
		Type: "DAMAGE",
		Data: map[string]interface{}{
			"playerId":    playerId,
			"srcBaseHero": srcBaseHero,
			"dstBaseHero": dstBaseHero,
			"damage":      damage,
		},
	})
	if err != nil {
		return err
	}

	// err = service.heroService.ApplyDamage(tx, actionLog.Order, srcHero, trgHero, damage)
	// if err != nil {
	// 	return err
	// }

	err = service.eventManager.Emit("DAMAGE", &AttackEvent{
		Tx:          tx,
		SrcHero:     srcHero,
		DstHero:     dstHero,
		Damage:      damage,
		CurrentTurn: session.CurrentTurn,
	})
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
