package matches

import (
	"errors"
	"strings"

	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/utils"
)

type IAction interface {
	apply(session *Session) error
	getSourcePlayerId() string
	getName() string
}

type MoveLog struct {
	PlayerId      string
	HeroName      string
	DirectionList []string
}

func (log MoveLog) apply(session *Session) error {
	hero, err := session.getHeroOnPlayer(log.PlayerId, log.HeroName)
	if err != nil {
		return err
	}

	if !hero.canMove(session.currentTurn) {
		return errors.New("hero already moved this turn")
	}

	if hero.Health == 0 {
		return exceptions.HeroIsDead()
	}

	if len(log.DirectionList) > hero.GetBaseStats().Range {
		return errors.New("invalid movement range")
	}

	curPos := hero.Pos
	for _, dir := range log.DirectionList {
		dirPoint := GetPointFromDirection(dir)
		curPos = curPos.Add(dirPoint)
		if !session.isPointOpen(curPos) {
			return errors.New("point is occupied")
		}
	}

	hero.Pos = curPos
	hero.lastMoveTurn = session.currentTurn
	return nil
}

func (log MoveLog) getSourcePlayerId() string {
	return log.PlayerId
}

func (log MoveLog) getName() string {
	return log.PlayerId + ": " + log.HeroName + " moved " + strings.Join(log.DirectionList, " ")
}

type AttackLog struct {
	SourcePlayerId string
	SourceHeroName string
	TargetPlayerId string
	TargetHeroName string
	Pos            Point
}

func (log AttackLog) apply(session *Session) error {
	srcHero, srcErr := session.getHeroOnPlayer(log.SourcePlayerId, log.SourceHeroName)
	if srcErr != nil {
		return srcErr
	}

	if !srcHero.canAttack(session.currentTurn) {
		return errors.New("hero already attacked this turn")
	}

	if srcHero.Health == 0 {
		return exceptions.HeroIsDead()
	}

	damage := srcHero.GetBaseStats().Damage
	srcHero.Health = max(srcHero.Health-damage, 0)
	srcHero.lastAttackTurn = session.currentTurn
	return nil
}

func (log AttackLog) getSourcePlayerId() string {
	return log.SourcePlayerId
}

func GetAction(actionName string, actionBody map[string]interface{}) (IAction, error) {
	if actionName == "move" {
		var action MoveLog
		err := utils.MapToObject(actionBody, &action)
		if err != nil {
			return nil, err
		}
		return action, nil
	} else if actionName == "attack" {
		var action AttackLog
		err := utils.MapToObject(actionBody, &action)
		if err != nil {
			return nil, err
		}
		return action, nil
	}
	return nil, errors.New("invalid action name")
}

func (log AttackLog) getName() string {
	return log.SourcePlayerId + ": " + log.SourceHeroName + " attacked " + log.TargetHeroName
}
