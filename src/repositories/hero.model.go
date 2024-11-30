package repositories

import (
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/utils/physics"
)

const (
	BASE_HERO_PREFIX = "hero_"
)

type HeroRepository interface{}

type HeroKey struct {
	SessionId string
	PlayerId  string
	BaseHero  string
}

type CreateHeroParams struct {
	Health         int
	Position       physics.Point
	LastMoveTurn   int
	LastAttackTurn int
	BaseHero       heroes.BaseHeroEnum

	PlayerId  string
	SessionId string
}

type HeroRepositoryImpl struct {
	badger databases.Badger
}

func (repo *HeroRepositoryImpl) GetHeroBySessionId(params HeroKey) *models.Hero {
	var hero models.Hero
	err := repo.badger.Get(BASE_HERO_PREFIX+params.SessionId+"_"+params.PlayerId+"_"+params.BaseHero, &hero)
	if err != nil {
		return nil
	}
	return &hero
}

func (repo *HeroRepositoryImpl) CreateHero(params CreateHeroParams) (*models.Hero, error) {
	hero := &models.Hero{
		Health:         params.Health,
		Position:       params.Position,
		LastMoveTurn:   params.LastMoveTurn,
		LastAttackTurn: params.LastAttackTurn,
		BaseHero:       params.BaseHero,
		PlayerId:       params.PlayerId,
		SessionId:      params.SessionId,
	}
	err := repo.badger.Set(BASE_HERO_PREFIX+params.SessionId+"_"+params.PlayerId+"_"+params.BaseHero, hero)
	if err != nil {
		return nil, err
	}
	return hero, nil
}

func (repo *HeroRepositoryImpl) DeleteHero(params HeroKey) error {
	var hero models.Hero
	err := repo.badger.Get(BASE_HERO_PREFIX+params.SessionId+"_"+params.PlayerId+"_"+params.BaseHero, &hero)
	if err != nil {
		return err
	}
	return repo.badger.Delete(BASE_HERO_PREFIX + params.SessionId + "_" + params.PlayerId + "_" + params.BaseHero)
}

func NewHeroRepositoryImpl(
	badger databases.Badger,
) *HeroRepositoryImpl {
	return &HeroRepositoryImpl{
		badger: badger,
	}
}
