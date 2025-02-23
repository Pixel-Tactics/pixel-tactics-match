package services

type ActionService interface {
	// GetSessionById(sessionId string) *models.Session
}

type ActionServiceImpl struct {
	// mapService     MapService
	// heroService    HeroService
	// sessionService SessionService

	// actionRepository   repositories.ActionLogRepository
	// transactionManager databases.TransactionManager
}

// func (service *ActionServiceImpl) Move(actionLog *models.ActionLog) error {
// 	session := service.sessionService.GetSessionById(actionLog.SessionId)
// 	if session == nil {
// 		return errors.New("invalid session id")
// 	}

// 	activePlayerId, err := session.GetActivePlayer()
// 	if err != nil {
// 		return err
// 	}

// 	action, ok := actionLog.Action.(*actions.MoveAction)
// 	if !ok {
// 		return errors.New("invalid action")
// 	}

// 	srcHero, err := service.heroService.GetPlayerHero(actionLog.SessionId, actionLog.PlayerId, action.SourceHero)
// 	if err != nil {
// 		return err
// 	}

// 	if !srcHero.CanMoveOnTurn(activePlayerId, actionLog.Turn) {
// 		return errors.New("hero cannot move")
// 	}

// 	srcHeroStats := service.heroService.GetStats(srcHero.BaseHero)
// 	if len(action.DirectionList) > srcHeroStats.BaseStats.MoveRange {
// 		return errors.New("invalid movement range")
// 	}

// 	curPos := srcHero.Position
// 	for _, dir := range action.DirectionList {
// 		dirPoint := physics.GetPointFromDirection(dir)
// 		curPos = curPos.Add(dirPoint)
// 		isPointOpen, err := service.mapService.IsPointOpen(session, curPos)
// 		if err != nil {
// 			return err
// 		}
// 		if !isPointOpen {
// 			return errors.New("point is occupied")
// 		}
// 	}

// 	tx := service.transactionManager.NewReadWriteTransaction()
// 	defer tx.Discard()

// 	service.heroService.MoveHero(tx, actionLog.Order, srcHero, curPos)

// 	err = tx.Commit()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (service *ActionServiceImpl) Attack(actionLog *models.ActionLog) error {
// 	session := service.sessionService.GetSessionById(actionLog.SessionId)
// 	if session == nil {
// 		return errors.New("invalid session id")
// 	}

// 	action, ok := actionLog.Action.(*actions.AttackAction)
// 	if !ok {
// 		return errors.New("invalid action")
// 	}

// 	srcHero, err := service.heroService.GetPlayerHero(actionLog.SessionId, actionLog.PlayerId, action.SourceHero)
// 	if err != nil {
// 		return err
// 	}

// 	otherPlayerId, _ := session.GetOtherPlayerId(actionLog.PlayerId)
// 	trgHero, err := service.heroService.GetPlayerHero(actionLog.SessionId, otherPlayerId, action.TargetHero)
// 	if err != nil {
// 		return err
// 	}

// 	activePlayerId, err := session.GetActivePlayer()
// 	if err != nil {
// 		return err
// 	}

// 	if !srcHero.CanAttackOnTurn(activePlayerId, actionLog.Turn) {
// 		return errors.New("hero cannot attack")
// 	}

// 	srcHeroStats := service.heroService.GetStats(srcHero.BaseHero)

// 	attackRange := srcHeroStats.BaseStats.AttackRange
// 	damage := srcHeroStats.BaseStats.Damage

// 	sessionMap, err := service.mapService.GetSessionMap(session.Id)
// 	if err != nil {
// 		return err
// 	}

// 	dist, err := algorithms.CheckDistance(sessionMap.Structure, srcHero.Position, trgHero.Position)
// 	if dist > attackRange || err != nil {
// 		return errors.New("target out of range")
// 	}

// 	tx := service.transactionManager.NewReadWriteTransaction()
// 	defer tx.Discard()

// 	action.Damage = damage
// 	_, err = service.actionRepository.UpdateActionLog(tx, actionLog)
// 	if err != nil {
// 		return err
// 	}
// 	err = service.heroService.ApplyDamage(tx, actionLog.Order, srcHero, trgHero, damage)
// 	if err != nil {
// 		return err
// 	}

// 	err = tx.Commit()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
