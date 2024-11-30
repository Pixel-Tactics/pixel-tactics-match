package heroes

type Mage struct{}

func (mage *Mage) GetInfo() *BaseHeroInfo {
	return &BaseHeroInfo{
		Name: BaseHeroMage,
		BaseStats: &BaseHeroStats{
			MaxHealth:   6,
			Damage:      2,
			AttackRange: 3,
			MoveRange:   2,
		},
	}
}

func NewMage() *Mage {
	return &Mage{}
}
