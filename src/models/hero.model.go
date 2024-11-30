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
	PlayerUsername string
	BaseHero       heroes.BaseHeroEnum
}
