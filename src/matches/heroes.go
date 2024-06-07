package matches

type Hero struct {
	Health         int   `json:"health"`
	Pos            Point `json:"pos"`
	lastMoveTurn   int
	lastAttackTurn int
	player         *Player
	HeroTemplate   `json:"template"`
}

func (h *Hero) canMove() bool {
	if h.Health == 0 {
		return false
	}
	if h.lastAttackTurn >= h.player.session.currentTurn {
		return false
	} else if h.lastMoveTurn >= h.player.session.currentTurn {
		return false
	} else {
		return true
	}
}

func (h *Hero) canAttack() bool {
	if h.Health == 0 {
		return false
	}
	if h.lastMoveTurn == h.player.session.currentTurn {
		action, avail := h.player.session.getLastAction()
		if avail {
			moveAction, ok := action.(*MoveLog)
			if ok && moveAction.srcHero != h {
				return false
			}
		}
	}
	return h.lastAttackTurn < h.player.session.currentTurn
}

func (h *Hero) getData() map[string]interface{} {
	return map[string]interface{}{
		"health":   h.Health,
		"pos":      h.Pos,
		"template": h.HeroTemplate.GetData(),
	}
}

func NewHero(template HeroTemplate, player *Player) *Hero {
	baseStats := template.GetBaseStats()
	return &Hero{
		Health:         baseStats.MaxHealth,
		HeroTemplate:   template,
		player:         player,
		lastMoveTurn:   -2,
		lastAttackTurn: -2,
	}
}
