package matches

import (
	"errors"
	"time"

	"pixeltactics.com/match/src/exceptions"
)

const (
	preparationTime = 90 * time.Second
	playerTurnTime  = 90 * time.Second
	numberOfHero    = 1
)

type ISessionState interface {
	preparePlayer(playerId string, chosenHeroList []*Hero) error
	startBattle() error
	executeAction(action IAction) error
	forfeit(playerId string) error
	getData() map[string]interface{}
}

// Preparation
type PreparationState struct {
	session  *Session
	deadline time.Time
}

func (state *PreparationState) executeAction(action IAction) error {
	return exceptions.ActionNotAllowed()
}

func (state *PreparationState) forfeit(playerId string) error {
	return exceptions.ActionNotAllowed()
}

func (state *PreparationState) preparePlayer(playerId string, chosenHeroList []*Hero) error {
	if time.Now().After(state.deadline) {
		state.session.state = &EndState{
			session:  state.session,
			winnerId: "draw",
		}
		state.session.processEndResult()
		return exceptions.ExceededDeadlineError()
	}

	player := state.session.getPlayerFromId(playerId)
	if player == nil {
		return errors.New("invalid player id")
	}

	if len(chosenHeroList) != numberOfHero {
		return errors.New("invalid number of heroes")
	}

	player.HeroList = chosenHeroList
	return nil
}

func (state *PreparationState) startBattle() error {
	if time.Now().After(state.deadline) {
		state.session.state = &EndState{
			session:  state.session,
			winnerId: "draw",
		}
		state.session.processEndResult()
		return exceptions.ExceededDeadlineError()
	}

	heroList1 := state.session.player1.HeroList
	heroList2 := state.session.player2.HeroList
	if len(heroList1)*len(heroList2) == 0 {
		return exceptions.HeroPickupError()
	}

	state.session.state = &Player1TurnState{
		session:  state.session,
		deadline: state.deadline.Add(playerTurnTime),
	}
	return nil
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

func (state *Player1TurnState) executeAction(action IAction) error {
	if time.Now().After(state.deadline) {
		state.session.state = &Player2TurnState{
			session:  state.session,
			deadline: state.deadline.Add(playerTurnTime),
		}
		return exceptions.ExceededDeadlineError()
	}

	playerId := action.getSourcePlayerId()
	if state.session.player1.Id != playerId {
		return exceptions.ActionNotAllowed()
	}

	err := action.apply(state.session)
	if err != nil {
		return err
	}

	state.session.actionLog = append(state.session.actionLog, action)

	winnerId := state.session.checkWinner()
	if winnerId != "" {
		state.session.state = &EndState{
			session:  state.session,
			winnerId: winnerId,
		}
		state.session.processEndResult()
	} else {
		state.session.state = &Player2TurnState{
			session:  state.session,
			deadline: state.deadline.Add(playerTurnTime),
		}
	}

	return nil
}

func (state *Player1TurnState) forfeit(playerId string) error {
	if state.session.player1.Id == playerId {
		state.session.state = &EndState{
			session:  state.session,
			winnerId: state.session.player2.Id,
		}
	} else {
		state.session.state = &EndState{
			session:  state.session,
			winnerId: state.session.player1.Id,
		}
	}
	state.session.processEndResult()
	return nil
}

func (state *Player1TurnState) preparePlayer(playerId string, chosenHeroList []*Hero) error {
	panic("unimplemented")
}

func (state *Player1TurnState) startBattle() error {
	panic("unimplemented")
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

func (state *Player2TurnState) executeAction(action IAction) error {
	if time.Now().After(state.deadline) {
		state.session.state = &Player1TurnState{
			session:  state.session,
			deadline: state.deadline.Add(playerTurnTime),
		}
		return exceptions.ExceededDeadlineError()
	}

	playerId := action.getSourcePlayerId()
	if state.session.player2.Id != playerId {
		return exceptions.ActionNotAllowed()
	}

	err := action.apply(state.session)
	if err != nil {
		return err
	}

	state.session.actionLog = append(state.session.actionLog, action)

	winnerId := state.session.checkWinner()
	if winnerId != "" {
		state.session.state = &EndState{
			session:  state.session,
			winnerId: winnerId,
		}
		state.session.processEndResult()
	} else {
		state.session.state = &Player1TurnState{
			session:  state.session,
			deadline: state.deadline.Add(playerTurnTime),
		}
	}

	return nil
}

func (state *Player2TurnState) forfeit(playerId string) error {
	if state.session.player1.Id == playerId {
		state.session.state = &EndState{
			session:  state.session,
			winnerId: state.session.player2.Id,
		}
	} else {
		state.session.state = &EndState{
			session:  state.session,
			winnerId: state.session.player1.Id,
		}
	}
	state.session.processEndResult()
	return nil
}

func (state *Player2TurnState) preparePlayer(playerId string, chosenHeroList []*Hero) error {
	panic("unimplemented")
}

func (state *Player2TurnState) startBattle() error {
	panic("unimplemented")
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

func (state *EndState) executeAction(action IAction) error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) forfeit(playerId string) error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) preparePlayer(playerId string, chosenHeroList []*Hero) error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) startBattle() error {
	return exceptions.ActionNotAllowed()
}

func (state *EndState) getData() map[string]interface{} {
	return map[string]interface{}{
		"name": "END",
	}
}
