package matches_interfaces

import matches_physics "pixeltactics.com/match/src/matches/physics"

type IHero interface {
	CanMove() bool
	CanAttack() bool
	GetPos() matches_physics.Point
	SetPos(newValue matches_physics.Point)
	SetLastAttackTurn(newValue int)
	SetLastMoveTurn(newValue int)
	GetHealth() int
	SetHealth(newValue int)
	GetPlayer() IPlayer

	HeroTemplate
}
