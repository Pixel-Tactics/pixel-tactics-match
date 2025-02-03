package services

import (
	"errors"

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
	CreateHeroes(sessionId string, playerId string, chosen []heroes.BaseHeroEnum) error
	InitHeroPosition(heroList1 []*models.Hero, spawnPoints1 []physics.Point, heroList2 []*models.Hero, spawnPoints2 []physics.Point) error
}

type HeroServiceImpl struct {
	heroRepository repositories.HeroRepository
	sessionService SessionService
	playerService  PlayerService

	heroFactory heroes.BaseHeroFactory
}

func (service *HeroServiceImpl) GetPlayerHeroes(sessionId string, playerId string) ([]*models.Hero, error) {
	player := service.playerService.GetPlayer(sessionId, playerId)
	if player == nil {
		return nil, errors.New("invalid player")
	}

	return service.heroRepository.GetPlayerHeroes(sessionId, playerId, player.HeroBases)
}

func (service *HeroServiceImpl) GetPlayerHero(sessionId string, playerId string, baseHero heroes.BaseHeroEnum) (*models.Hero, error) {
	heroes, err := service.GetPlayerHeroes(sessionId, playerId)
	if err != nil {
		return nil, err
	}

	for _, hero := range heroes {
		if hero.BaseHero == baseHero {
			return hero, nil
		}
	}
	return nil, errors.New("hero not found")
}

func (service *HeroServiceImpl) GetStats(heroName heroes.BaseHeroEnum) *heroes.BaseHeroInfo {
	hero := service.heroFactory.Create(heroName)
	return hero.GetInfo()
}

func (service *HeroServiceImpl) GetAvailableHeroes() []heroes.BaseHeroEnum {
	return []heroes.BaseHeroEnum{
		heroes.BaseHeroKnight,
		heroes.BaseHeroMage,
	}
}

func (service *HeroServiceImpl) CreateHeroes(sessionId string, playerId string, chosen []heroes.BaseHeroEnum) error {
	session := service.sessionService.GetSessionById(sessionId)
	if session == nil {
		return exceptions.SessionNotFound()
	}

	isValid := service.isChosenHeroesValid(session.AllowedHeroList, chosen)
	if !isValid {
		return exceptions.HeroPickupError()
	}

	params := []repositories.CreateHeroParams{}
	for _, heroEnum := range chosen {
		baseHero := service.heroFactory.Create(heroEnum)
		baseHeroInfo := baseHero.GetInfo()
		params = append(params, repositories.CreateHeroParams{
			Health:         baseHeroInfo.BaseStats.MaxHealth,
			LastMoveTurn:   -2,
			LastAttackTurn: -2,
			BaseHero:       heroEnum,
			PlayerId:       playerId,
			SessionId:      session.Id,
		})
	}

	err := service.playerService.SetPlayerHeroes(sessionId, playerId, chosen)
	if err != nil {
		return err
	}

	_, err = service.heroRepository.BatchCreateHeroes(params)
	if err != nil {
		service.playerService.SetPlayerHeroes(sessionId, playerId, make([]heroes.BaseHeroEnum, 0))
		return err
	}

	return nil
}

func (service *HeroServiceImpl) InitHeroPosition(
	heroList1 []*models.Hero,
	spawnPoints1 []physics.Point,
	heroList2 []*models.Hero,
	spawnPoints2 []physics.Point,
) error {
	updateParams := make([]repositories.UpdateHeroParams, 0)
	for i, spawnPoint := range spawnPoints1 {
		if i < len(heroList1) {
			updateParams = append(updateParams, repositories.UpdateHeroParams{
				Health:         heroList1[i].Health,
				LastMoveTurn:   heroList1[i].LastMoveTurn,
				LastAttackTurn: heroList1[i].LastAttackTurn,
				Position:       spawnPoint,
				SessionId:      heroList1[i].SessionId,
				PlayerId:       heroList1[i].PlayerId,
				BaseHero:       heroList1[i].BaseHero,
			})
		} else {
			break
		}
	}

	for i, spawnPoint := range spawnPoints2 {
		if i < len(heroList2) {
			updateParams = append(updateParams, repositories.UpdateHeroParams{
				Health:         heroList2[i].Health,
				LastMoveTurn:   heroList2[i].LastMoveTurn,
				LastAttackTurn: heroList2[i].LastAttackTurn,
				Position:       spawnPoint,
				SessionId:      heroList2[i].SessionId,
				PlayerId:       heroList2[i].PlayerId,
				BaseHero:       heroList2[i].BaseHero,
			})
		} else {
			break
		}
	}

	return service.heroRepository.BatchUpdateHero(updateParams)
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
