package matches_heroes_templates

import matches_interfaces "pixeltactics.com/match/src/matches/interfaces"

type Knight struct{}

func (k *Knight) GetBaseStats() matches_interfaces.HeroTemplateStats {
	return matches_interfaces.HeroTemplateStats{
		MaxHealth:   10,
		Damage:      4,
		AttackRange: 1,
		MoveRange:   3,
	}
}

func (k *Knight) GetName() string {
	return "knight"
}

func (k *Knight) GetData() map[string]interface{} {
	return map[string]interface{}{
		"name":  k.GetName(),
		"stats": k.GetBaseStats(),
	}
}
