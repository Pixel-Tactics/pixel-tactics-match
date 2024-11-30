package heroes

type BaseHeroFactory interface {
	Create(name BaseHeroEnum) BaseHero
}

type BaseHeroFactoryImpl struct{}

func (factory *BaseHeroFactoryImpl) Create(name BaseHeroEnum) BaseHero {
	if name == BaseHeroKnight {
		return NewKnight()
	} else {
		return NewMage()
	}
}
