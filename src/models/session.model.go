package models

type SessionState = string

const (
	SessionStateMatchMaking = "SessionStateMatchMaking"
	SessionStatePreparation = "SessionStatePreparation"
	SessionStatePlayer1Turn = "SessionStatePlayer1Turn"
	SessionStatePlayer2Turn = "SessionStatePlayer2Turn"
	SessionStateEnded       = "SessionStateEnded"
)

type Session struct {
	Id             string
	Turn           int
	State          SessionState
	HeroList       []string
	PlayerUsername []string
}

func (session *Session) IsRunning() bool {
	return session.State == SessionStateMatchMaking || session.State == SessionStateEnded
}
