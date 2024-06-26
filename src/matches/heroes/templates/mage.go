package matches_heroes_templates

import matches_interfaces "pixeltactics.com/match/src/matches/interfaces"

type Mage struct{}

func (m *Mage) GetBaseStats() matches_interfaces.HeroTemplateStats {
	return matches_interfaces.HeroTemplateStats{
		MaxHealth:   6,
		Damage:      2,
		AttackRange: 3,
		MoveRange:   2,
	}
}

func (m *Mage) GetName() string {
	return "mage"
}

func (m *Mage) GetData() map[string]interface{} {
	return map[string]interface{}{
		"name":  m.GetName(),
		"stats": m.GetBaseStats(),
	}
}
