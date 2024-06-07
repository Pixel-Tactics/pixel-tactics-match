package matches

type Player struct {
	Id       string
	HeroList []*Hero
	session  *Session
}

func (p *Player) IsHeroExists(hero *Hero) bool {
	for _, curHero := range p.HeroList {
		if hero == curHero {
			return true
		}
	}
	return false
}

func (p *Player) hasAvailableAction() bool {
	result := false
	for _, hero := range p.HeroList {
		if hero.canAttack() {
			result = true
			break
		}
	}
	return result
}

func (p *Player) getData() map[string]interface{} {
	var heroListData = []map[string]interface{}{}
	for _, hero := range p.HeroList {
		heroListData = append(heroListData, hero.getData())
	}
	return map[string]interface{}{
		"id":       p.Id,
		"heroList": heroListData,
	}
}

func NewPlayer(id string, session *Session) *Player {
	return &Player{
		Id:       id,
		HeroList: []*Hero{},
		session:  session,
	}
}
