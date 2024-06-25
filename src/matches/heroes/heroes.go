package matches_heroes

import (
	matches_constants "pixeltactics.com/match/src/matches/constants"
	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
	matches_physics "pixeltactics.com/match/src/matches/physics"
)

type Hero struct {
	Health                          int                   `json:"health"`
	Pos                             matches_physics.Point `json:"pos"`
	lastMoveTurn                    int
	lastAttackTurn                  int
	player                          matches_interfaces.IPlayer
	matches_interfaces.HeroTemplate `json:"template"`
}

func (h *Hero) CanMove() bool {
	if h.Health == 0 {
		return false
	}
	if h.lastAttackTurn >= h.player.GetSession().GetCurrentTurn() {
		return false
	} else if h.lastMoveTurn >= h.player.GetSession().GetCurrentTurn() {
		return false
	} else {
		return true
	}
}

func (h *Hero) CanAttack() bool {
	if h.Health == 0 {
		return false
	}
	if h.lastMoveTurn == h.player.GetSession().GetCurrentTurn() {
		action, avail := h.player.GetSession().GetLastAction()
		if avail && action.GetName() == matches_constants.MOVE_LOG {
			sourceHero, ok := action.GetSourceHero().(*Hero)
			if ok && sourceHero != h {
				return false
			}
		}
	}
	return h.lastAttackTurn < h.player.GetSession().GetCurrentTurn()
}

func (h *Hero) GetData() map[string]interface{} {
	return map[string]interface{}{
		"health":   h.Health,
		"pos":      h.Pos,
		"template": h.HeroTemplate.GetData(),
	}
}

func (h *Hero) GetHealth() int {
	return h.Health
}

func (h *Hero) SetHealth(newValue int) {
	h.Health = newValue
}

func (h *Hero) GetPlayer() matches_interfaces.IPlayer {
	return h.player
}

func (h *Hero) GetPos() matches_physics.Point {
	return h.Pos
}

func (h *Hero) SetPos(newValue matches_physics.Point) {
	h.Pos = newValue
}

func (h *Hero) SetLastAttackTurn(newValue int) {
	h.lastAttackTurn = newValue
}

func (h *Hero) SetLastMoveTurn(newValue int) {
	h.lastMoveTurn = newValue
}

func NewHero(template matches_interfaces.HeroTemplate, player matches_interfaces.IPlayer) *Hero {
	baseStats := template.GetBaseStats()
	return &Hero{
		Health:         baseStats.MaxHealth,
		HeroTemplate:   template,
		player:         player,
		lastMoveTurn:   -2,
		lastAttackTurn: -2,
	}
}
