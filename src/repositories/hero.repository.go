package repositories

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/utils/physics"
)

const (
	BASE_HERO_PREFIX = "hero_"
)

type HeroRepository interface {
	GetPlayerHeroes(sessionId string, playerId string, bases []heroes.BaseHeroEnum) ([]*models.Hero, error)
	GetHeroBySessionId(params HeroKey) *models.Hero
	CreateHero(params CreateHeroParams) (*models.Hero, error)
	UpdateHero(params UpdateHeroParams) (*models.Hero, error)
	BatchUpdateHero(params []UpdateHeroParams) error
	BatchCreateHeroes(params []CreateHeroParams) ([]*models.Hero, error)
	DeleteHero(params HeroKey) error
}

type HeroKey struct {
	SessionId string
	PlayerId  string
	BaseHero  string
}

type UpdateHeroParams struct {
	Health         int
	Position       physics.Point
	LastMoveTurn   int
	LastAttackTurn int

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

func (repo *HeroRepositoryImpl) GetPlayerHeroes(sessionId string, playerId string, bases []heroes.BaseHeroEnum) ([]*models.Hero, error) {
	ret := make([]*models.Hero, 0)
	for _, base := range bases {
		var hero models.Hero
		err := repo.badger.Get(BASE_HERO_PREFIX+sessionId+"_"+playerId+"_"+base, &hero)
		if err != nil {
			return nil, errors.New("invalid params")
		}
		ret = append(ret, &hero)
	}
	return ret, nil
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

func (repo *HeroRepositoryImpl) UpdateHero(params UpdateHeroParams) (*models.Hero, error) {
	hero := repo.GetHeroBySessionId(HeroKey{
		SessionId: params.SessionId,
		PlayerId:  params.PlayerId,
		BaseHero:  params.BaseHero,
	})
	if hero == nil {
		return nil, errors.New("invalid hero key")
	}

	hero.Health = params.Health
	hero.Position = params.Position
	hero.LastMoveTurn = params.LastMoveTurn
	hero.LastAttackTurn = params.LastAttackTurn

	err := repo.badger.Set(BASE_HERO_PREFIX+params.SessionId+"_"+params.PlayerId+"_"+params.BaseHero, hero)
	if err != nil {
		return nil, err
	}
	return hero, nil
}

func (repo *HeroRepositoryImpl) BatchUpdateHero(params []UpdateHeroParams) error {
	setParams := make([]*databases.BadgerSetParams, 0)
	for _, param := range params {
		hero := repo.GetHeroBySessionId(HeroKey{
			SessionId: param.SessionId,
			PlayerId:  param.PlayerId,
			BaseHero:  param.BaseHero,
		})
		if hero == nil {
			return errors.New("invalid hero key")
		}
		hero.Health = param.Health
		hero.Position = param.Position
		hero.LastAttackTurn = param.LastAttackTurn
		hero.LastMoveTurn = param.LastMoveTurn
		setParams = append(setParams, &databases.BadgerSetParams{
			Key:   BASE_HERO_PREFIX + hero.SessionId + "_" + hero.PlayerId + "_" + hero.BaseHero,
			Value: hero,
		})
	}

	return repo.badger.BatchSet(setParams)
}

func (repo *HeroRepositoryImpl) BatchCreateHeroes(params []CreateHeroParams) ([]*models.Hero, error) {
	ret := make([]*models.Hero, 0)
	setParams := make([]*databases.BadgerSetParams, 0)
	for _, param := range params {
		hero := &models.Hero{
			Health:         param.Health,
			Position:       param.Position,
			LastMoveTurn:   param.LastMoveTurn,
			LastAttackTurn: param.LastAttackTurn,
			BaseHero:       param.BaseHero,
			PlayerId:       param.PlayerId,
			SessionId:      param.SessionId,
		}
		ret = append(ret, hero)
		setParams = append(setParams, &databases.BadgerSetParams{
			Key:   BASE_HERO_PREFIX + param.SessionId + "_" + param.PlayerId + "_" + param.BaseHero,
			Value: hero,
		})
	}
	err := repo.badger.BatchSet(setParams)
	if err != nil {
		return nil, err
	}
	return ret, nil
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
