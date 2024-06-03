package matches

type HeroTemplateStats struct {
	MaxHealth int
	Damage    int
	Range     int
}

type HeroTemplate interface {
	GetBaseStats() HeroTemplateStats
	GetName() string
}

type Knight struct{}

// GetStats implements HeroTemplate.
func (k Knight) GetBaseStats() HeroTemplateStats {
	return HeroTemplateStats{
		MaxHealth: 10,
		Damage:    4,
		Range:     1,
	}
}

func (k Knight) GetName() string {
	return "knight"
}

type Mage struct{}

func (m Mage) GetBaseStats() HeroTemplateStats {
	return HeroTemplateStats{
		MaxHealth: 6,
		Damage:    2,
		Range:     3,
	}
}

func (m Mage) GetName() string {
	return "mage"
}
