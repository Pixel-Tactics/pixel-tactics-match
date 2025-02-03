package models

import (
	"pixeltactics.com/match/src/core/actions"
)

type ActionLog struct {
	SessionId string
	PlayerId  string
	Order     int

	Turn   int
	Action actions.Action
}
