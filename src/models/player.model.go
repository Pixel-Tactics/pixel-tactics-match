package models

import "pixeltactics.com/match/src/heroes"

type Player struct {
	Id        string
	SessionId string
	HeroBases []heroes.BaseHeroEnum
}
