package matches

type HeroTemplateStats struct {
	MaxHealth   int `json:"maxHealth"`
	Damage      int `json:"damage"`
	AttackRange int `json:"attackRange"`
	MoveRange   int `json:"moveRange"`
}

type HeroTemplate interface {
	GetBaseStats() HeroTemplateStats
	GetName() string
	GetData() map[string]interface{}
}

type Knight struct{}

// GetStats implements HeroTemplate.
func (k *Knight) GetBaseStats() HeroTemplateStats {
	return HeroTemplateStats{
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

type Mage struct{}

func (m *Mage) GetBaseStats() HeroTemplateStats {
	return HeroTemplateStats{
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

func GetAvailableHeroes() ([]string, error) {
	return []string{"knight"}, nil
}

func Test() map[string]interface{} {
	test := &Mage{}
	return test.GetData()
}
