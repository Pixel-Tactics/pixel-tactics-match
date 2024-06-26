package matches_actions

import (
	"errors"

	matches_algorithms "pixeltactics.com/match/src/matches/algorithms"
	matches_constants "pixeltactics.com/match/src/matches/constants"
	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
)

type AttackLog struct {
	srcHero matches_interfaces.IHero
	trgHero matches_interfaces.IHero
	damage  int
}

type AttackLogData struct {
	HeroName   string `json:"heroName"`
	PlayerId   string `json:"playerId"`
	TargetName string `json:"targetName"`
}

func (log *AttackLog) Apply(session matches_interfaces.ISession) error {
	if !log.srcHero.CanAttack() {
		return errors.New("hero cannot attack")
	}

	attackRange := log.srcHero.GetBaseStats().AttackRange
	damage := log.srcHero.GetBaseStats().Damage

	dist, err := matches_algorithms.CheckDistance(session.GetMatchMap().Structure, log.srcHero.GetPos(), log.trgHero.GetPos())
	if dist > attackRange || err != nil {
		return errors.New("target out of range")
	}

	log.damage = damage
	log.srcHero.SetLastAttackTurn(session.GetCurrentTurn())
	log.trgHero.SetHealth(max(log.trgHero.GetHealth()-damage, 0))
	return nil
}

func (log *AttackLog) GetSourcePlayerId() string {
	return log.srcHero.GetPlayer().GetId()
}

func (log *AttackLog) GetData() map[string]interface{} {
	return map[string]interface{}{
		"hero":     log.srcHero.GetName(),
		"target":   log.trgHero.GetName(),
		"playerId": log.srcHero.GetPlayer().GetId(),
		"damage":   log.damage,
	}
}

func (log *AttackLog) GetName() string {
	return matches_constants.ATTACK_LOG
}

func (log *AttackLog) GetSourceHero() matches_interfaces.IHero {
	return log.srcHero
}

func NewAttackLog(srcHero matches_interfaces.IHero, trgHero matches_interfaces.IHero) *AttackLog {
	return &AttackLog{
		srcHero: srcHero,
		trgHero: trgHero,
	}
}
