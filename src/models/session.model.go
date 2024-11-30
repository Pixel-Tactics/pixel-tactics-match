package models

import (
	"errors"
	"time"

	"pixeltactics.com/match/src/heroes"
)

type SessionState = string

const (
	SessionStateMatchMaking = "SessionStateMatchMaking"
	SessionStatePreparation = "SessionStatePreparation"
	SessionStatePlayer1Turn = "SessionStatePlayer1Turn"
	SessionStatePlayer2Turn = "SessionStatePlayer2Turn"
	SessionStateEnded       = "SessionStateEnded"
)

type Session struct {
	Id              string
	State           State
	AllowedHeroList []heroes.BaseHeroEnum
	PlayerIds       []string
	WinnerId        *string
}

type State struct {
	Id       string
	Type     SessionState
	Deadline time.Time
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
