package matches

import (
	"errors"
	"time"

	"pixeltactics.com/match/src/exceptions"
	matches_heroes "pixeltactics.com/match/src/matches/heroes"
	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
	matches_physics "pixeltactics.com/match/src/matches/physics"
	"pixeltactics.com/match/src/utils"
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
	expire()
}

// Preparation
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

func (state *PreparationState) getData() map[string]interface{} {
	return map[string]interface{}{
		"name":     "PREPARATION",
		"deadline": float64(state.deadline.UnixMilli()) / 1000.0,
	}
}

// Player 1 Turn
type Player1TurnState struct {
	session  *Session
	deadline time.Time
}

func (state *Player1TurnState) executeAction(action matches_interfaces.IAction) error {
	if time.Now().After(state.deadline) {
		return exceptions.ExceededDeadlineError()
	}

	playerId := action.GetSourcePlayerId()
	if state.session.player1.Id != playerId {
		return exceptions.ActionNotAllowed()
	}

	err := action.Apply(state.session)
	if err != nil {
		return err
	}

	state.session.actionLog = append(state.session.actionLog, action)

	winnerId := state.session.checkWinner()
	if winnerId != "" {
		state.session.changeState(&EndState{
			session:  state.session,
			winnerId: winnerId,
		})
		return nil
	}

	hasAction := state.session.player1.HasAvailableAction()
	if !hasAction {
		state.session.changeState(&Player2TurnState{
			session:  state.session,
			deadline: time.Now().Add(playerTurnTime),
		})
	}

	return nil
}

func (state *Player1TurnState) endTurn(playerId string) error {
	if state.session.player1.Id == playerId {
		state.session.changeState(&Player2TurnState{
			session:  state.session,
			deadline: utils.MinTime(time.Now(), state.deadline).Add(playerTurnTime),
		})
		return nil
	} else {
		return exceptions.ActionNotAllowed()
	}
}

func (state *Player1TurnState) forfeit(playerId string) error {
	if state.session.player1.Id == playerId {
		state.session.changeState(&EndState{
			session:  state.session,
			winnerId: state.session.player2.Id,
		})
	} else {
		state.session.changeState(&EndState{
			session:  state.session,
			winnerId: state.session.player1.Id,
		})
	}
	return nil
}

func (state *Player1TurnState) preparePlayer(playerId string, chosenHeroList []matches_interfaces.HeroTemplate) error {
	return exceptions.ActionNotAllowed()
}

func (state *Player1TurnState) startBattle() error {
	return exceptions.ActionNotAllowed()
}

func (state *Player1TurnState) expire() {
	state.session.lock.Lock()
	defer state.session.lock.Unlock()
	if state.session.state == state {
		state.session.changeState(&Player2TurnState{
			session:  state.session,
			deadline: state.deadline.Add(playerTurnTime),
		})
	}
}

func (state *Player1TurnState) getData() map[string]interface{} {
	return map[string]interface{}{
		"name":     "PLAYER_1_TURN",
		"deadline": float64(state.deadline.UnixMilli()) / 1000.0,
	}
}

// Player 2 Turn
type Player2TurnState struct {
	session  *Session
	deadline time.Time
}

func (state *Player2TurnState) executeAction(action matches_interfaces.IAction) error {
	if time.Now().After(state.deadline) {
		return exceptions.ExceededDeadlineError()
	}

	playerId := action.GetSourcePlayerId()
	if state.session.player2.Id != playerId {
		return exceptions.ActionNotAllowed()
	}

	err := action.Apply(state.session)
	if err != nil {
		return err
	}

	state.session.actionLog = append(state.session.actionLog, action)

	winnerId := state.session.checkWinner()
	if winnerId != "" {
		state.session.changeState(&EndState{
			session:  state.session,
			winnerId: winnerId,
		})
		return nil
	}

	hasAction := state.session.player2.HasAvailableAction()
	if !hasAction {
		state.session.changeState(&Player1TurnState{
			session:  state.session,
			deadline: time.Now().Add(playerTurnTime),
		})
	}

	return nil
}

func (state *Player2TurnState) endTurn(playerId string) error {
	if state.session.player2.Id == playerId {
		state.session.changeState(&Player1TurnState{
			session:  state.session,
			deadline: utils.MinTime(time.Now(), state.deadline).Add(playerTurnTime),
		})
		return nil
	} else {
		return exceptions.ActionNotAllowed()
	}
}

func (state *Player2TurnState) forfeit(playerId string) error {
	if state.session.player1.Id == playerId {
		state.session.changeState(&EndState{
			session:  state.session,
			winnerId: state.session.player2.Id,
		})
	} else {
		state.session.changeState(&EndState{
			session:  state.session,
			winnerId: state.session.player1.Id,
		})
	}
	return nil
}

func (state *Player2TurnState) preparePlayer(playerId string, chosenHeroList []matches_interfaces.HeroTemplate) error {
	return exceptions.ActionNotAllowed()
}

func (state *Player2TurnState) startBattle() error {
	return exceptions.ActionNotAllowed()
}

func (state *Player2TurnState) expire() {
	state.session.lock.Lock()
	defer state.session.lock.Unlock()
	if state.session.state == state {
		state.session.changeState(&Player1TurnState{
			session:  state.session,
			deadline: state.deadline.Add(playerTurnTime),
		})
	}
}

func (state *Player2TurnState) getData() map[string]interface{} {
	return map[string]interface{}{
		"name":     "PLAYER_2_TURN",
		"deadline": float64(state.deadline.UnixMilli()) / 1000.0,
	}
}

// End
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

func (state *EndState) expire() {}

func (state *EndState) getData() map[string]interface{} {
	return map[string]interface{}{
		"name":     "END",
		"winnerId": state.winnerId,
	}
}
