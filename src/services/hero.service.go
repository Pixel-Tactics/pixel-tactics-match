package services

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
	"pixeltactics.com/match/src/utils/physics"
)

const (
	NumberOfHero = 1
)

type HeroService interface {
	GetStats(heroName heroes.BaseHeroEnum) *heroes.BaseHeroInfo
	GetPlayerHeroes(sessionId string, playerId string) ([]*models.Hero, error)
	GetPlayerHero(sessionId string, playerId string, baseHero heroes.BaseHeroEnum) (*models.Hero, error)
	GetAvailableHeroes() []heroes.BaseHeroEnum
	CreateHeroesTx(tx databases.BadgerTx, sessionId string, playerId string, chosen []heroes.BaseHeroEnum) error
	CreateHeroes(sessionId string, playerId string, chosen []heroes.BaseHeroEnum) error
	InitHeroPosition(heroList1 []*models.Hero, spawnPoints1 []physics.Point, heroList2 []*models.Hero, spawnPoints2 []physics.Point) error

	ApplyDamage(tx databases.BadgerTx, currentTurn int, srcHero *models.Hero, trgHero *models.Hero, damage int) error
	MoveHero(tx databases.BadgerTx, currentTurn int, srcHero *models.Hero, position physics.Point) error
}

type HeroServiceImpl struct {
	heroRepository repositories.HeroRepository
	sessionService SessionService
	playerService  PlayerService

	heroFactory        heroes.BaseHeroFactory
	transactionManager databases.TransactionManager
}

func (service *HeroServiceImpl) GetPlayerHeroes(sessionId string, playerId string) ([]*models.Hero, error) {
	player := service.playerService.GetPlayer(sessionId, playerId)
	if player == nil {
		return nil, errors.New("invalid player")
	}

	return service.heroRepository.GetPlayerHeroes(nil, sessionId, playerId, player.HeroBases)
}

func (service *HeroServiceImpl) GetPlayerHero(sessionId string, playerId string, baseHero heroes.BaseHeroEnum) *models.Hero {
	return service.heroRepository.GetHeroBySessionId(nil, repositories.HeroKey{
		SessionId: sessionId,
		PlayerId:  playerId,
		BaseHero:  baseHero,
	})
}

func (service *HeroServiceImpl) GetStats(heroName heroes.BaseHeroEnum) *heroes.BaseHeroInfo {
	hero := service.heroFactory.Create(heroName)
	return hero.GetInfo()
}

func (service *HeroServiceImpl) MoveHero(tx databases.BadgerTx, currentTurn int, srcHero *models.Hero, position physics.Point) error {
	srcHero.Position = position
	srcHero.LastMoveTurn = currentTurn

	_, err := service.heroRepository.SaveHero(tx, srcHero)
	if err != nil {
		return err
	}

	return nil
}

func (service *HeroServiceImpl) ApplyDamage(tx databases.BadgerTx, currentTurn int, srcHero *models.Hero, trgHero *models.Hero, damage int) error {
	srcHero.LastAttackTurn = currentTurn
	trgHero.Health = max(trgHero.Health-damage, 0)

	_, err := service.heroRepository.SaveHero(tx, srcHero)
	if err != nil {
		return err
	}
	_, err = service.heroRepository.SaveHero(tx, trgHero)
	if err != nil {
		return err
	}

	return nil
}

func (service *HeroServiceImpl) GetAvailableHeroes() []heroes.BaseHeroEnum {
	return []heroes.BaseHeroEnum{
		heroes.BaseHeroKnight,
		heroes.BaseHeroMage,
	}
}

func (service *HeroServiceImpl) CreateHeroesTx(tx databases.BadgerTx, sessionId string, playerId string, chosen []heroes.BaseHeroEnum) error {
	session := service.sessionService.GetSessionById(sessionId)
	if session == nil {
		return exceptions.SessionNotFound()
	}

	isValid := service.isChosenHeroesValid(session.AllowedHeroList, chosen)
	if !isValid {
		return exceptions.HeroPickupError()
	}

	for _, heroEnum := range chosen {
		baseHero := service.heroFactory.Create(heroEnum)
		baseHeroInfo := baseHero.GetInfo()
		_, err := service.heroRepository.SaveHero(tx, &models.Hero{
			Health:         baseHeroInfo.BaseStats.MaxHealth,
			LastMoveTurn:   -2,
			LastAttackTurn: -2,
			BaseHero:       heroEnum,
			PlayerId:       playerId,
			SessionId:      session.Id,
		})
		if err != nil {
			return err
		}
	}

	err := service.playerService.SetPlayerHeroes(tx, sessionId, playerId, chosen)
	if err != nil {
		return err
	}

	return nil
}

func (service *HeroServiceImpl) CreateHeroes(sessionId string, playerId string, chosen []heroes.BaseHeroEnum) error {
	tx := service.transactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	err := service.CreateHeroesTx(tx, sessionId, playerId, chosen)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (service *HeroServiceImpl) InitHeroPositionTx(
	tx databases.BadgerTx,
	heroList1 []*models.Hero,
	spawnPoints1 []physics.Point,
	heroList2 []*models.Hero,
	spawnPoints2 []physics.Point,
) error {

	for i, spawnPoint := range spawnPoints1 {
		if i < len(heroList1) {
			heroList1[i].Position = spawnPoint
			_, err := service.heroRepository.SaveHero(tx, heroList1[i])
			if err != nil {
				return err
			}
		} else {
			break
		}
	}

	for i, spawnPoint := range spawnPoints2 {
		if i < len(heroList2) {
			heroList2[i].Position = spawnPoint
			_, err := service.heroRepository.SaveHero(tx, heroList2[i])
			if err != nil {
				return err
			}
		} else {
			break
		}
	}

	return tx.Commit()
}

func (service *HeroServiceImpl) InitHeroPosition(
	heroList1 []*models.Hero,
	spawnPoints1 []physics.Point,
	heroList2 []*models.Hero,
	spawnPoints2 []physics.Point,
) error {
	tx := service.transactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	err := service.InitHeroPositionTx(tx, heroList1, spawnPoints1, heroList2, spawnPoints2)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (service *HeroServiceImpl) isChosenHeroesValid(available []heroes.BaseHeroEnum, chosen []heroes.BaseHeroEnum) bool {
	if len(chosen) != NumberOfHero {
		return false
	}

	dupeCheck := make(map[heroes.BaseHeroEnum]bool)
	for _, hero := range chosen {
		// Check if hero is available
		exists := false
		for _, availHero := range available {
			if availHero == hero {
				exists = true
				break
			}
		}
		if !exists {
			return false
		}

		// Check for duplication
		_, dupeExists := dupeCheck[hero]
		if dupeExists {
			return false
		} else {
			dupeCheck[hero] = true
		}
	}
	return true
}
