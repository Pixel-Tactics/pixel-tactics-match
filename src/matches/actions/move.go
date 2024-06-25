package matches_actions

import (
	"errors"

	"pixeltactics.com/match/src/exceptions"
	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
	matches_physics "pixeltactics.com/match/src/matches/physics"
)

type MoveLog struct {
	srcHero       matches_interfaces.IHero
	directionList []string
}

type MoveLogData struct {
	HeroName      string   `json:"heroName"`
	PlayerId      string   `json:"playerId"`
	DirectionList []string `json:"directionList"`
}

func (log *MoveLog) Apply(session matches_interfaces.ISession) error {
	if !log.srcHero.CanMove() {
		return errors.New("hero already moved this turn")
	}

	if log.srcHero.GetHealth() == 0 {
		return exceptions.HeroIsDead()
	}

	if len(log.directionList) > log.srcHero.GetBaseStats().MoveRange {
		return errors.New("invalid movement range")
	}

	curPos := log.srcHero.GetPos()
	for _, dir := range log.directionList {
		dirPoint := matches_physics.GetPointFromDirection(dir)
		curPos = curPos.Add(dirPoint)
		if !session.IsPointOpen(curPos) {
			return errors.New("point is occupied")
		}
	}

	log.srcHero.SetPos(curPos)
	log.srcHero.SetLastMoveTurn(session.GetCurrentTurn())
	return nil
}

func (log *MoveLog) GetSourcePlayerId() string {
	return log.srcHero.GetPlayer().GetId()
}

func (log *MoveLog) GetData() map[string]interface{} {
	return map[string]interface{}{
		"hero":          log.srcHero.GetName(),
		"playerId":      log.srcHero.GetPlayer().GetId(),
		"directionList": log.directionList,
	}
}

func (log *MoveLog) GetName() string {
	return "move"
}

func (log *MoveLog) GetSourceHero() matches_interfaces.IHero {
	return log.srcHero
}

func NewMoveLog(hero matches_interfaces.IHero, directionList []string) *MoveLog {
	return &MoveLog{
		srcHero:       hero,
		directionList: directionList,
	}
}
