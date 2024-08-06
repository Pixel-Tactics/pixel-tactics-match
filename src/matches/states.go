package matches

import (
	"time"

	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
)

const (
	preparationTime = 60 * time.Second
	playerTurnTime  = 30 * time.Second
	numberOfHero    = 1
)

type ISessionState interface {
	preparePlayer(playerId string, chosenHeroList []matches_interfaces.HeroTemplate) error
	startBattle() error
	executeAction(action matches_interfaces.IAction) error
	endTurn(playerId string) error
	forfeit(playerId string) error
	getData() map[string]interface{}
}

type ITimedState interface {
	getDeadline() time.Time
	expire()
}
