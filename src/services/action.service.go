package services

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
	"pixeltactics.com/match/src/utils/algorithms"
	convert_utils "pixeltactics.com/match/src/utils/convert"
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
}

type MoveSessionLog struct {
	PlayerId string
	BaseHero string
	Point    physics.Point
}

func (service *ActionServiceImpl) ApplyMoveLog(log MoveSessionLog, currentTurn int, srcHero *models.Hero) {
	srcHero.MovePosition(currentTurn, log.Point)
}

func (service *ActionServiceImpl) Move(sessionId string, playerId string, baseHero string, directions []physics.Direction) error {
	tx := service.transactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	session, err := service.sessionService.GetSessionById(tx, sessionId)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("invalid session id")
	}

	activePlayerId, err := session.GetActivePlayer()
	if err != nil {
		return err
	}

	srcHero, err := service.heroService.GetPlayerHero(tx, sessionId, playerId, baseHero)
	if err != nil {
		return err
	}
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

	logObj, err := convert_utils.ObjectToMap(&MoveSessionLog{
		PlayerId: playerId,
		BaseHero: baseHero,
		Point:    curPos,
	})
	if err != nil {
		return err
	}

	_, err = service.sessionLogRepository.AppendLog(tx, sessionId, &models.SessionLog{
		Type: "MOVE",
		Data: logObj,
	})
	if err != nil {
		return err
	}

	err = service.heroService.MoveHero(tx, session.CurrentTurn, srcHero, curPos)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (service *ActionServiceImpl) ApplyAttackLog(log AttackSessionLog, currentTurn int, srcHero *models.Hero, dstHero *models.Hero) {
	srcHero.Attack(currentTurn)
	dstHero.Damage(log.Damage)
}

type AttackSessionLog struct {
	PlayerId    string
	SrcBaseHero string
	DstBaseHero string
	Damage      int
}

func (service *ActionServiceImpl) Attack(sessionId string, playerId string, srcBaseHero string, dstBaseHero string) error {
	tx := service.transactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	session, err := service.sessionService.GetSessionById(tx, sessionId)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("invalid session id")
	}

	srcHero, err := service.heroService.GetPlayerHero(tx, sessionId, playerId, srcBaseHero)
	if err != nil {
		return err
	}
	if srcHero == nil {
		return errors.New("invalid source hero")
	}

	otherPlayerId, _ := session.GetOtherPlayerId(playerId)
	dstHero, err := service.heroService.GetPlayerHero(tx, sessionId, otherPlayerId, dstBaseHero)
	if err != nil {
		return err
	}
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

	logObj, err := convert_utils.ObjectToMap(&AttackSessionLog{
		PlayerId:    playerId,
		SrcBaseHero: srcBaseHero,
		DstBaseHero: dstBaseHero,
		Damage:      damage,
	})
	if err != nil {
		return err
	}

	_, err = service.sessionLogRepository.AppendLog(tx, sessionId, &models.SessionLog{
		Type: "DAMAGE",
		Data: logObj,
	})
	if err != nil {
		return err
	}

	err = service.heroService.ApplyDamage(tx, session.CurrentTurn, srcHero, dstHero, damage)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
