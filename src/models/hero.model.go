package models

import (
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/utils/physics"
)

// LastMoveTurn >= LastAttackTurn
type Hero struct {
	Health         int           `json:"health"`
	Position       physics.Point `json:"pos"`
	LastMoveTurn   int
	LastAttackTurn int
	BaseHero       heroes.BaseHeroEnum

	PlayerId  string
	SessionId string
}

func (hero *Hero) MovePosition(currentTurn int, position physics.Point) {
	hero.Position = position
	hero.LastMoveTurn = currentTurn
}

func (hero *Hero) Attack(currentTurn int) {
	hero.LastAttackTurn = currentTurn
}

func (hero *Hero) Damage(damage int) {
	hero.Health = max(hero.Health-damage, 0)
}

func (hero *Hero) CanMoveOnTurn(curPlayerId string, currentTurn int) bool {
	if hero.Health == 0 || curPlayerId != hero.PlayerId {
		return false
	}
	return currentTurn > hero.LastMoveTurn
}

func (hero *Hero) CanAttackOnTurn(curPlayerId string, currentTurn int) bool {
	if hero.Health == 0 || curPlayerId != hero.PlayerId {
		return false
	}
	return currentTurn == hero.LastMoveTurn && currentTurn > hero.LastAttackTurn
}
