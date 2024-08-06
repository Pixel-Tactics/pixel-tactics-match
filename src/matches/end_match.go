package matches

import (
	"pixeltactics.com/match/src/exceptions"
	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
)

type EndState struct {
	session  *Session
	winnerId string
}

func (state *EndState) executeAction(action matches_interfaces.IAction) error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) endTurn(playerId string) error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) forfeit(playerId string) error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) preparePlayer(playerId string, chosenHeroList []matches_interfaces.HeroTemplate) error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) startBattle() error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) getData() map[string]interface{} {
	return map[string]interface{}{
		"name":     "END",
		"winnerId": state.winnerId,
	}
}
