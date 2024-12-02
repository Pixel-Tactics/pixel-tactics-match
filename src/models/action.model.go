package models

import (
	"pixeltactics.com/match/src/core/actions"
)

type ActionLog struct {
	SessionId string
	PlayerId  string
	Order     int

	Action actions.Action
}
