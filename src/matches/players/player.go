package matches_players

import matches_interfaces "pixeltactics.com/match/src/matches/interfaces"

type Player struct {
	Id       string
	HeroList []matches_interfaces.IHero
	session  matches_interfaces.ISession
}

func (p *Player) IsHeroExists(hero matches_interfaces.IHero) bool {
	for _, curHero := range p.HeroList {
		if hero == curHero {
			return true
		}
	}
	return false
}

func (p *Player) HasAvailableAction() bool {
	result := false
	for _, hero := range p.HeroList {
		if hero.CanAttack() {
			result = true
			break
		}
	}
	return result
}

func (p *Player) GetData() map[string]interface{} {
	var heroListData = []map[string]interface{}{}
	for _, hero := range p.HeroList {
		heroListData = append(heroListData, hero.GetData())
	}
	return map[string]interface{}{
		"id":       p.Id,
		"heroList": heroListData,
	}
}

func (p *Player) GetId() string {
	return p.Id
}

func (p *Player) GetSession() matches_interfaces.ISession {
	return p.session
}

func (p *Player) GetHeroList() []matches_interfaces.IHero {
	return p.HeroList
}

func NewPlayer(id string, heroList []matches_interfaces.IHero, session matches_interfaces.ISession) *Player {
	return &Player{
		Id:       id,
		HeroList: heroList,
		session:  session,
	}
}
