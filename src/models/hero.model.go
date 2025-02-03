package models

import (
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/utils/physics"
)

type Hero struct {
	Health         int           `json:"health"`
	Position       physics.Point `json:"pos"`
	LastMoveTurn   int
	LastAttackTurn int
	BaseHero       heroes.BaseHeroEnum

	PlayerId  string
	SessionId string
}

func (hero *Hero) CanAttackOnTurn(curPlayerId string, currentTurn int) bool {
	// LastMoveTurn >= LastAttackTurn
	if curPlayerId != hero.PlayerId {
		return false
	}
	return currentTurn == hero.LastMoveTurn && currentTurn > hero.LastAttackTurn
}
