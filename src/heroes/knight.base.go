package heroes

type Knight struct{}

func (knight *Knight) GetInfo() *BaseHeroInfo {
	return &BaseHeroInfo{
		Name: BaseHeroKnight,
		BaseStats: &BaseHeroStats{
			MaxHealth:   10,
			Damage:      4,
			AttackRange: 1,
			MoveRange:   3,
		},
	}
}

func NewKnight() *Knight {
	return &Knight{}
}
