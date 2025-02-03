package repositories

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/models"
)

const (
	BASE_HERO_PREFIX = "hero_"
)

type HeroRepository interface {
	GetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string, bases []heroes.BaseHeroEnum) ([]*models.Hero, error)
	GetHeroBySessionId(tx databases.BadgerTx, params HeroKey) *models.Hero
	SaveHero(tx databases.BadgerTx, params *models.Hero) (*models.Hero, error)
	// UpdateHero(params UpdateHeroParams) (*models.Hero, error)
	// BatchUpdateHero(params []UpdateHeroParams) error
	// BatchCreateHeroes(params []CreateHeroParams) ([]*models.Hero, error)
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

func (repo *HeroRepositoryImpl) GetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string, bases []heroes.BaseHeroEnum) ([]*models.Hero, error) {
	query := databases.GetQuery(tx, repo.badger)

	ret := make([]*models.Hero, 0)
	for _, base := range bases {
		var hero models.Hero
		err := query.Get(BASE_HERO_PREFIX+sessionId+"_"+playerId+"_"+base, &hero)
		if err != nil {
			return nil, errors.New("invalid params")
		}
		ret = append(ret, &hero)
	}
	return ret, nil
}

func (repo *HeroRepositoryImpl) GetHeroBySessionId(tx databases.BadgerTx, params HeroKey) *models.Hero {
	query := databases.GetQuery(tx, repo.badger)

	var hero models.Hero
	err := query.Get(BASE_HERO_PREFIX+params.SessionId+"_"+params.PlayerId+"_"+params.BaseHero, &hero)
	if err != nil {
		return nil
	}
	return &hero
}

func (repo *HeroRepositoryImpl) SaveHero(tx databases.BadgerTx, obj *models.Hero) (*models.Hero, error) {
	err := tx.Set(BASE_HERO_PREFIX+obj.SessionId+"_"+obj.PlayerId+"_"+obj.BaseHero, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// func (repo *HeroRepositoryImpl) UpdateHero(obj *models.Hero) (*models.Hero, error) {
// 	// hero := repo.GetHeroBySessionId(HeroKey{
// 	// 	SessionId: obj.SessionId,
// 	// 	PlayerId:  obj.PlayerId,
// 	// 	BaseHero:  obj.BaseHero,
// 	// })
// 	// if hero == nil {
// 	// 	return nil, errors.New("invalid hero key")
// 	// }

// 	err := repo.badger.Set(BASE_HERO_PREFIX+params.SessionId+"_"+params.PlayerId+"_"+params.BaseHero, hero)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return hero, nil
// }

// func (repo *HeroRepositoryImpl) BatchUpdateHero(params []UpdateHeroParams) error {
// 	setParams := make([]*databases.BadgerSetParams, 0)
// 	for _, param := range params {
// 		hero := repo.GetHeroBySessionId(HeroKey{
// 			SessionId: param.SessionId,
// 			PlayerId:  param.PlayerId,
// 			BaseHero:  param.BaseHero,
// 		})
// 		if hero == nil {
// 			return errors.New("invalid hero key")
// 		}
// 		if param.Health != nil {
// 			hero.Health = *param.Health
// 		}
// 		if param.Position != nil {
// 			hero.Position = *param.Position
// 		}
// 		if param.LastMoveTurn != nil {
// 			hero.LastMoveTurn = *param.LastMoveTurn
// 		}
// 		if param.LastAttackTurn != nil {
// 			hero.LastAttackTurn = *param.LastAttackTurn
// 		}
// 		setParams = append(setParams, &databases.BadgerSetParams{
// 			Key:   BASE_HERO_PREFIX + hero.SessionId + "_" + hero.PlayerId + "_" + hero.BaseHero,
// 			Value: hero,
// 		})
// 	}

// 	return repo.badger.BatchSet(setParams)
// }

// func (repo *HeroRepositoryImpl) BatchCreateHeroes(params []CreateHeroParams) ([]*models.Hero, error) {
// 	ret := make([]*models.Hero, 0)
// 	setParams := make([]*databases.BadgerSetParams, 0)
// 	for _, param := range params {
// 		hero := &models.Hero{
// 			Health:         param.Health,
// 			Position:       param.Position,
// 			LastMoveTurn:   param.LastMoveTurn,
// 			LastAttackTurn: param.LastAttackTurn,
// 			BaseHero:       param.BaseHero,
// 			PlayerId:       param.PlayerId,
// 			SessionId:      param.SessionId,
// 		}
// 		ret = append(ret, hero)
// 		setParams = append(setParams, &databases.BadgerSetParams{
// 			Key:   BASE_HERO_PREFIX + param.SessionId + "_" + param.PlayerId + "_" + param.BaseHero,
// 			Value: hero,
// 		})
// 	}
// 	err := repo.badger.BatchSet(setParams)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return ret, nil
// }

func (repo *HeroRepositoryImpl) DeleteHero(tx databases.BadgerTx, params HeroKey) error {
	var hero models.Hero
	err := tx.Get(BASE_HERO_PREFIX+params.SessionId+"_"+params.PlayerId+"_"+params.BaseHero, &hero)
	if err != nil {
		return err
	}
	return tx.Delete(BASE_HERO_PREFIX + params.SessionId + "_" + params.PlayerId + "_" + params.BaseHero)
}

func NewHeroRepositoryImpl(
	badger databases.Badger,
) *HeroRepositoryImpl {
	return &HeroRepositoryImpl{
		badger: badger,
	}
}
