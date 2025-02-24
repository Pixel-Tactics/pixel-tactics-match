package models

import (
	"errors"
	"time"

	"pixeltactics.com/match/src/heroes"
)

type SessionState = string

const (
	SessionStateMatchMaking = "MATCH_MAKING"
	SessionStatePreparation = "PREPARATION"
	SessionStatePlayer1Turn = "PLAYER_1_TURN"
	SessionStatePlayer2Turn = "PLAYER_2_TURN"
	SessionStateEnded       = "ENDED"
)

type Session struct {
	Id              string
	State           State
	AllowedHeroList []heroes.BaseHeroEnum
	PlayerIds       []string
	WinnerId        *string
}

type State struct {
	Id       string       `json:"id"`
	Type     SessionState `json:"name"`
	Deadline time.Time    `json:"deadline"`
}

func (session *Session) IsRunning() bool {
	return session.State.Type != SessionStateMatchMaking
}

func (session *Session) GetOtherPlayerId(playerId string) (string, error) {
	if playerId != session.PlayerIds[0] && playerId != session.PlayerIds[1] {
		return "", errors.New("invalid player id")
	}
	if playerId == session.PlayerIds[0] {
		return session.PlayerIds[1], nil
	} else {
		return session.PlayerIds[0], nil
	}
}

func (session *Session) GetActivePlayer() (string, error) {
	if session.State.Type == SessionStatePlayer1Turn {
		return session.PlayerIds[0], nil
	} else if session.State.Type == SessionStatePlayer2Turn {
		return session.PlayerIds[0], nil
	}
	return "", errors.New("no player is active")
}
