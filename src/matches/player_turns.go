package matches

import (
	"time"

	"pixeltactics.com/match/src/exceptions"
	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
	"pixeltactics.com/match/src/notifiers"
	"pixeltactics.com/match/src/utils"
)

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

	actionData := action.GetData()
	actionData["order"] = len(state.session.actionLog) - 1

	notifier := notifiers.GetSessionNotifier()
	notifier.NotifyAction(state.session.player1.Id, action.GetName(), actionData)
	notifier.NotifyAction(state.session.player2.Id, action.GetName(), actionData)

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

func (state *Player1TurnState) getDeadline() time.Time {
	return state.deadline
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

	actionData := action.GetData()
	actionData["order"] = len(state.session.actionLog) - 1

	notifier := notifiers.GetSessionNotifier()
	notifier.NotifyAction(state.session.player1.Id, action.GetName(), actionData)
	notifier.NotifyAction(state.session.player2.Id, action.GetName(), actionData)

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

func (state *Player2TurnState) getDeadline() time.Time {
	return state.deadline
}
