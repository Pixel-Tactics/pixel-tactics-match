package services

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/utils/algorithms"
	convert_utils "pixeltactics.com/match/src/utils/convert"
	"pixeltactics.com/match/src/utils/physics"
)

type ActionService interface {
	Move(playerId string, baseHero string, directions []physics.Direction) error
	Attack(playerId string, srcBaseHero string, dstBaseHero string) error
	EndTurn(playerId string) error
}

type ActionServiceImpl struct {
	MapService     MapService
	HeroService    HeroService
	SessionService SessionService

	LogService         LogService
	TransactionManager databases.TransactionManager
}

type MoveSessionLog struct {
	PlayerId string
	BaseHero string
	Point    physics.Point
}

func (service *ActionServiceImpl) ApplyMoveLog(log MoveSessionLog, currentTurn int, srcHero *models.Hero) {
	srcHero.MovePosition(currentTurn, log.Point)
}

func (service *ActionServiceImpl) Move(playerId string, baseHero string, directions []physics.Direction) error {
	tx := service.TransactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	session, err := service.getSessionIfPlayerTurn(tx, playerId)
	if err != nil {
		return err
	}

	srcHero, err := service.HeroService.GetPlayerHero(tx, session.Id, playerId, baseHero)
	if err != nil {
		return err
	}
	if srcHero == nil {
		return errors.New("invalid hero")
	}

	if !srcHero.CanMoveOnTurn(playerId, session.CurrentTurn) {
		return errors.New("hero cannot move")
	}

	srcHeroStats := service.HeroService.GetStats(srcHero.BaseHero)
	if len(directions) > srcHeroStats.BaseStats.MoveRange {
		return errors.New("invalid movement range")
	}

	curPos := srcHero.Position
	for _, dir := range directions {
		dirPoint := physics.GetPointFromDirection(dir)
		curPos = curPos.Add(dirPoint)
		isPointOpen, err := service.MapService.IsPointOpen(tx, session, curPos)
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

	err = service.LogService.AppendLog(tx, session, &models.SessionLog{
		Type:      "MOVE",
		SessionId: session.Id,
		Data:      logObj,
	})
	if err != nil {
		return err
	}

	err = service.HeroService.MoveHero(tx, session.CurrentTurn, srcHero, curPos)
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

func (service *ActionServiceImpl) Attack(playerId string, srcBaseHero string, dstBaseHero string) error {
	tx := service.TransactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	session, err := service.getSessionIfPlayerTurn(tx, playerId)
	if err != nil {
		return err
	}

	srcHero, err := service.HeroService.GetPlayerHero(tx, session.Id, playerId, srcBaseHero)
	if err != nil {
		return err
	}
	if srcHero == nil {
		return errors.New("invalid source hero")
	}

	otherPlayerId, _ := session.GetOtherPlayerId(playerId)
	dstHero, err := service.HeroService.GetPlayerHero(tx, session.Id, otherPlayerId, dstBaseHero)
	if err != nil {
		return err
	}
	if dstHero == nil {
		return errors.New("invalid destination hero")
	}

	if !srcHero.CanAttackOnTurn(playerId, session.CurrentTurn) {
		return errors.New("hero cannot attack")
	}

	srcHeroStats := service.HeroService.GetStats(srcHero.BaseHero)

	attackRange := srcHeroStats.BaseStats.AttackRange
	damage := srcHeroStats.BaseStats.Damage

	sessionMap, err := service.MapService.GetSessionMap(tx, session.Id)
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

	err = service.LogService.AppendLog(tx, session, &models.SessionLog{
		Type:      "DAMAGE",
		SessionId: session.Id,
		Data:      logObj,
	})
	if err != nil {
		return err
	}

	err = service.HeroService.ApplyDamage(tx, session.CurrentTurn, srcHero, dstHero, damage)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (service *ActionServiceImpl) EndTurn(playerId string) error {
	tx := service.TransactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	session, err := service.getSessionIfPlayerTurn(tx, playerId)
	if err != nil {
		return err
	}

	err = service.SessionService.SwapTurn(tx, session)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (service *ActionServiceImpl) getSessionIfPlayerTurn(tx databases.BadgerTx, playerId string) (*models.Session, error) {
	session, err := service.SessionService.GetSessionByPlayerId(tx, playerId)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	activePlayerId, err := session.GetActivePlayer()
	if err != nil {
		return nil, err
	}

	if playerId != activePlayerId {
		return nil, ErrNotTurn
	}

	return session, nil
}

func NewActionService(
	mapService MapService,
	heroService HeroService,
	sessionService SessionService,
	logService LogService,
	transactionManager databases.TransactionManager,
) ActionService {
	return &ActionServiceImpl{
		MapService:         mapService,
		HeroService:        heroService,
		SessionService:     sessionService,
		LogService:         logService,
		TransactionManager: transactionManager,
	}
}
