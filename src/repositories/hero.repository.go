package repositories

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_HERO_PREFIX         = "hero_"
	BASE_HERO_INITIAL_PREFIX = "initial_hero_"
)

type HeroRepository interface {
	// When no key found (or empty) for heroes, or somehow atleast one hero is not found, error will be returned.
	GetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string, bases []heroes.BaseHeroEnum) ([]*models.Hero, error)
	// When no key found (or empty) for hero, nil hero will be returned instead of error.
	GetHeroBySessionId(tx databases.BadgerTx, params HeroKey) (*models.Hero, error)
	SaveHero(tx databases.BadgerTx, params *models.Hero) (*models.Hero, error)
	SaveHeroes(tx databases.BadgerTx, heroList []*models.Hero) ([]*models.Hero, error)
	GetInitialHero(tx databases.BadgerTx, sessionId string, playerId string, baseHero string) (*models.Hero, error)
	SaveInitialHero(tx databases.BadgerTx, obj *models.Hero) (*models.Hero, error)
	DeleteHero(tx databases.BadgerTx, params HeroKey) error
}

type HeroKey struct {
	SessionId string
	PlayerId  string
	BaseHero  string
}

type HeroRepositoryImpl struct {
	badger databases.Badger
}

// When no key found (or empty) for heroes, or somehow atleast one hero is not found, error will be returned.
func (repo *HeroRepositoryImpl) GetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string, bases []heroes.BaseHeroEnum) ([]*models.Hero, error) {
	query := databases.GetQuery(tx, repo.badger)

	ret := make([]*models.Hero, 0)
	for _, base := range bases {
		var hero models.Hero
		err := query.Get(BASE_HERO_PREFIX+sessionId+"_"+playerId+"_"+base, &hero)
		if err != nil && err == databases.NotFoundException() {
			return nil, errors.New("hero " + base + " of " + playerId + " not found")
		}
		if err != nil {
			return nil, err
		}
		ret = append(ret, &hero)
	}
	return ret, nil
}

// When no key found (or empty) for hero, nil hero will be returned instead of error.
func (repo *HeroRepositoryImpl) GetHeroBySessionId(tx databases.BadgerTx, params HeroKey) (*models.Hero, error) {
	query := databases.GetQuery(tx, repo.badger)

	var hero models.Hero
	err := query.Get(BASE_HERO_PREFIX+params.SessionId+"_"+params.PlayerId+"_"+params.BaseHero, &hero)
	if err != nil && err == databases.NotFoundException() {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &hero, nil
}

func (repo *HeroRepositoryImpl) SaveHero(tx databases.BadgerTx, obj *models.Hero) (*models.Hero, error) {
	query := databases.GetQuery(tx, repo.badger)

	err := query.Set(BASE_HERO_PREFIX+obj.SessionId+"_"+obj.PlayerId+"_"+obj.BaseHero, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (repo *HeroRepositoryImpl) SaveHeroes(tx databases.BadgerTx, heroList []*models.Hero) ([]*models.Hero, error) {
	if tx == nil {
		return nil, errors.New("transaction object is null")
	}

	for _, hero := range heroList {
		_, err := repo.SaveHero(tx, hero)
		if err != nil {
			return nil, err
		}
	}
	return heroList, nil
}

func (repo *HeroRepositoryImpl) GetInitialHero(tx databases.BadgerTx, sessionId string, playerId string, baseHero string) (*models.Hero, error) {
	query := databases.GetQuery(tx, repo.badger)

	var hero models.Hero
	err := query.Get(BASE_HERO_INITIAL_PREFIX+sessionId+"_"+playerId+"_"+baseHero, &hero)
	if err != nil && err == databases.NotFoundException() {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &hero, nil
}

func (repo *HeroRepositoryImpl) SaveInitialHero(tx databases.BadgerTx, obj *models.Hero) (*models.Hero, error) {
	query := databases.GetQuery(tx, repo.badger)

	err := query.Set(BASE_HERO_INITIAL_PREFIX+obj.SessionId+"_"+obj.PlayerId+"_"+obj.BaseHero, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (repo *HeroRepositoryImpl) DeleteHero(tx databases.BadgerTx, params HeroKey) error {
	if tx == nil {
		return errors.New("transaction object is null")
	}

	var hero models.Hero
	err := tx.Get(BASE_HERO_PREFIX+params.SessionId+"_"+params.PlayerId+"_"+params.BaseHero, &hero)
	if err != nil {
		return err
	}
	err = tx.Delete(BASE_HERO_PREFIX + params.SessionId + "_" + params.PlayerId + "_" + params.BaseHero)
	if err != nil {
		return err
	}
	return tx.Delete(BASE_HERO_INITIAL_PREFIX + params.SessionId + "_" + params.PlayerId + "_" + params.BaseHero)
}

func NewHeroRepository(
	badger databases.Badger,
) HeroRepository {
	return &HeroRepositoryImpl{
		badger: badger,
	}
}
