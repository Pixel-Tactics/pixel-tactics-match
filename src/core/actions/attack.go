package actions

import (
	"pixeltactics.com/match/src/heroes"
)

const (
	ActionTypeAttack ActionType = "ActionTypeAttack"
)

type AttackAction struct {
	SourcePlayerId string
	SourceHero     heroes.BaseHeroEnum
	TargetHero     heroes.BaseHeroEnum
	Damage         int
}

func (action *AttackAction) GetType() ActionType {
	return ActionTypeAttack
}

func NewAttackAction(
	sourcePlayerId string,
	sourceHero heroes.BaseHeroEnum,
	targetHero heroes.BaseHeroEnum,
	damage int,
) *AttackAction {
	return &AttackAction{
		SourcePlayerId: sourcePlayerId,
		SourceHero:     sourceHero,
		TargetHero:     targetHero,
		Damage:         damage,
	}
}
