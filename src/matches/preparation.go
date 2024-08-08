package matches

import (
	"errors"
	"time"

	"pixeltactics.com/match/src/exceptions"
	matches_heroes "pixeltactics.com/match/src/matches/heroes"
	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
	matches_physics "pixeltactics.com/match/src/matches/physics"
)

type PreparationState struct {
	session  *Session
	deadline time.Time
}

func (state *PreparationState) executeAction(action matches_interfaces.IAction) error {
	return exceptions.ActionNotAllowed()
}

func (state *PreparationState) endTurn(playerId string) error {
	return exceptions.ActionNotAllowed()
}

func (state *PreparationState) forfeit(playerId string) error {
	return exceptions.ActionNotAllowed()
}

func (state *PreparationState) preparePlayer(playerId string, chosenHeroList []matches_interfaces.HeroTemplate) error {
	if time.Now().After(state.deadline) {
		return exceptions.ExceededDeadlineError()
	}

	player := state.session.getPlayerFromId(playerId)
	if player == nil {
		return errors.New("invalid player id")
	}

	heroList := []*matches_heroes.Hero{}
	for _, template := range chosenHeroList {
		heroList = append(heroList, matches_heroes.NewHero(template, player))
	}

	if len(chosenHeroList) != numberOfHero {
		return errors.New("invalid number of heroes")
	}

	iheroList := make([]matches_interfaces.IHero, len(heroList))
	for i := range heroList {
		iheroList[i] = heroList[i]
	}

	player.HeroList = iheroList
	return nil
}

func (state *PreparationState) startBattle() error {
	if time.Now().After(state.deadline) {
		return exceptions.ExceededDeadlineError()
	}

	heroList1 := state.session.player1.HeroList
	heroList2 := state.session.player2.HeroList
	if len(heroList1)*len(heroList2) == 0 {
		return exceptions.HeroPickupError()
	}

	mapStructure := state.session.matchMap.Structure
	if len(mapStructure) == 0 {
		return errors.New("invalid map structure")
	}

	cnt1 := 0
	cnt2 := 0
	for i := range mapStructure {
		for j := range mapStructure[0] {
			if mapStructure[i][j] == 3 && cnt1 < len(heroList1) {
				heroList1[cnt1].SetPos(matches_physics.Point{X: j, Y: i})
				cnt1 += 1
			} else if mapStructure[i][j] == 4 && cnt2 < len(heroList2) {
				heroList2[cnt2].SetPos(matches_physics.Point{X: j, Y: i})
				cnt2 += 1
			}
		}
	}

	state.session.changeState(&Player1TurnState{
		session:  state.session,
		deadline: time.Now().Add(playerTurnTime),
	})

	return nil
}

func (state *PreparationState) expire() {
	state.session.lock.Lock()
	defer state.session.lock.Unlock()
	if state.session.state == state {
		state.session.changeState(&EndState{
			session:  state.session,
			winnerId: "draw",
		})
	}
}

func (state *PreparationState) getDeadline() time.Time {
	return state.deadline
}

func (state *PreparationState) getData() map[string]interface{} {
	return map[string]interface{}{
		"name":     "PREPARATION",
		"deadline": float64(state.deadline.UnixMilli()) / 1000.0,
	}
}
