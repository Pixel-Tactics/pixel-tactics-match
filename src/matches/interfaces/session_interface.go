package matches_interfaces

import (
	matches_maps "pixeltactics.com/match/src/matches/maps"
	matches_physics "pixeltactics.com/match/src/matches/physics"
)

type ISession interface {
	IsPointOpen(point matches_physics.Point) bool
	GetCurrentTurn() int
	GetMatchMap() *matches_maps.MatchMap
	GetLastAction() (IAction, bool)
}
