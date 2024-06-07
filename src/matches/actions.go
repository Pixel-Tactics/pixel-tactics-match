package matches

import (
	"errors"

	"pixeltactics.com/match/src/exceptions"
)

type IAction interface {
	apply(session *Session) error
	getSourcePlayerId() string
	getData() map[string]interface{}
	getName() string
}

type MoveLog struct {
	srcHero       *Hero
	directionList []string
}

type MoveLogData struct {
	HeroName      string   `json:"heroName"`
	PlayerId      string   `json:"playerId"`
	DirectionList []string `json:"directionList"`
}

func (log *MoveLog) apply(session *Session) error {
	if !log.srcHero.canMove() {
		return errors.New("hero already moved this turn")
	}

	if log.srcHero.Health == 0 {
		return exceptions.HeroIsDead()
	}

	if len(log.directionList) > log.srcHero.GetBaseStats().MoveRange {
		return errors.New("invalid movement range")
	}

	curPos := log.srcHero.Pos
	for _, dir := range log.directionList {
		dirPoint := GetPointFromDirection(dir)
		curPos = curPos.Add(dirPoint)
		if !session.isPointOpen(curPos) {
			return errors.New("point is occupied")
		}
	}

	log.srcHero.Pos = curPos
	log.srcHero.lastMoveTurn = session.currentTurn
	return nil
}

func (log *MoveLog) getSourcePlayerId() string {
	return log.srcHero.player.Id
}

func (log *MoveLog) getData() map[string]interface{} {
	return map[string]interface{}{
		"hero":          log.srcHero.GetName(),
		"playerId":      log.srcHero.player.Id,
		"directionList": log.directionList,
	}
}

func (log *MoveLog) getName() string {
	return "move"
}

type AttackLog struct {
	srcHero *Hero
	trgHero *Hero
	damage  int
}

type AttackLogData struct {
	HeroName   string `json:"heroName"`
	PlayerId   string `json:"playerId"`
	TargetName string `json:"targetName"`
}

func (log *AttackLog) apply(session *Session) error {
	if !log.srcHero.canAttack() {
		return errors.New("hero cannot attack")
	}

	attackRange := log.srcHero.GetBaseStats().AttackRange
	damage := log.srcHero.GetBaseStats().Damage

	dist, err := CheckDistance(session.matchMap.Structure, log.srcHero.Pos, log.trgHero.Pos)
	if dist > attackRange || err != nil {
		return errors.New("target out of range")
	}

	log.damage = damage
	log.srcHero.lastAttackTurn = session.currentTurn
	log.trgHero.Health = max(log.trgHero.Health-damage, 0)
	return nil
}

func (log *AttackLog) getSourcePlayerId() string {
	return log.srcHero.player.Id
}

func (log *AttackLog) getData() map[string]interface{} {
	return map[string]interface{}{
		"hero":     log.srcHero.GetName(),
		"target":   log.trgHero.GetName(),
		"playerId": log.srcHero.player.Id,
		"damage":   log.damage,
	}
}

func (log *AttackLog) getName() string {
	return "attack"
}
