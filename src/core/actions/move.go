package actions

import (
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/utils/physics"
)

const (
	ActionTypeMove ActionType = "ActionTypeMove"
)

type MoveAction struct {
	SourcePlayerId string
	SourceHero     heroes.BaseHeroEnum
	DirectionList  []physics.Direction
}

func (action *MoveAction) GetType() ActionType {
	return ActionTypeMove
}

func NewMoveAction(
	sourcePlayerId string,
	sourceHero heroes.BaseHeroEnum,
	directionList []physics.Direction,
) *MoveAction {
	return &MoveAction{
		SourcePlayerId: sourcePlayerId,
		SourceHero:     sourceHero,
		DirectionList:  directionList,
	}
}
