package services

import (
	"errors"
	"log"

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

	// Error will be returned when heroes array doesn't exist.
	GetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string) ([]*models.Hero, error)

	// Nil will be returned when hero doesn't exist.
	GetPlayerHero(tx databases.BadgerTx, sessionId string, playerId string, baseHero heroes.BaseHeroEnum) (*models.Hero, error)

	GetAvailableHeroes() []heroes.BaseHeroEnum

	// Sets heroes for specific player. The chosen heroes must be valid in terms of number, availability, and duplication.
	CreateHeroes(tx databases.BadgerTx, sessionId string, playerId string, chosen []heroes.BaseHeroEnum) ([]*models.Hero, error)
	InitHeroPosition(tx databases.BadgerTx, heroList1 []*models.Hero, spawnPoints1 []physics.Point, heroList2 []*models.Hero, spawnPoints2 []physics.Point) error
	GetInitialState(tx databases.BadgerTx, heroList1 []*models.Hero, heroList2 []*models.Hero) ([]*models.Hero, []*models.Hero, error)

	ApplyDamage(tx databases.BadgerTx, currentTurn int, srcHero *models.Hero, trgHero *models.Hero, damage int) error
	MoveHero(tx databases.BadgerTx, currentTurn int, srcHero *models.Hero, position physics.Point) error

	SetSessionService(sessionService SessionService)
	SetPlayerService(playerService PlayerService)
	SetHeroes(tx databases.BadgerTx, heroList []*models.Hero) error
}

type HeroServiceImpl struct {
	HeroRepository repositories.HeroRepository
	LogRepository  repositories.SessionLogRepository
	SessionService SessionService
	PlayerService  PlayerService

	HeroFactory        heroes.BaseHeroFactory
	TransactionManager databases.TransactionManager
}

// Error will be returned when heroes array doesn't exist.
func (service *HeroServiceImpl) GetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string) ([]*models.Hero, error) {
	player, err := service.PlayerService.GetPlayer(tx, playerId, sessionId)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return nil, errors.New("invalid player")
	}

	if len(player.HeroBases) == 0 {
		return nil, ErrEmptyHero
	}

	heroes, err := service.HeroRepository.GetPlayerHeroes(nil, sessionId, playerId, player.HeroBases)
	if err != nil {
		return nil, err
	}
	return heroes, err
}

// Nil will be returned when hero doesn't exist.
func (service *HeroServiceImpl) GetPlayerHero(tx databases.BadgerTx, sessionId string, playerId string, baseHero heroes.BaseHeroEnum) (*models.Hero, error) {
	hero, err := service.HeroRepository.GetHeroBySessionId(nil, repositories.HeroKey{
		SessionId: sessionId,
		PlayerId:  playerId,
		BaseHero:  baseHero,
	})
	if err != nil {
		return nil, err
	}
	if hero == nil {
		return nil, nil
	}
	return hero, err
}

func (service *HeroServiceImpl) GetStats(heroName heroes.BaseHeroEnum) *heroes.BaseHeroInfo {
	hero := service.HeroFactory.Create(heroName)
	return hero.GetInfo()
}

func (service *HeroServiceImpl) MoveHero(tx databases.BadgerTx, currentTurn int, srcHero *models.Hero, position physics.Point) error {
	srcHero.MovePosition(currentTurn, position)

	_, err := service.HeroRepository.SaveHero(tx, srcHero)
	if err != nil {
		return err
	}

	return nil
}

func (service *HeroServiceImpl) ApplyDamage(tx databases.BadgerTx, currentTurn int, srcHero *models.Hero, trgHero *models.Hero, damage int) error {
	srcHero.Attack(currentTurn)
	trgHero.Damage(damage)

	_, err := service.HeroRepository.SaveHero(tx, srcHero)
	if err != nil {
		return err
	}
	_, err = service.HeroRepository.SaveHero(tx, trgHero)
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

// Sets heroes for specific player. The chosen heroes must be valid in terms of number, availability, and duplication.
func (service *HeroServiceImpl) CreateHeroes(tx databases.BadgerTx, sessionId string, playerId string, chosen []heroes.BaseHeroEnum) ([]*models.Hero, error) {
	session, err := service.SessionService.GetSessionById(tx, sessionId)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, exceptions.SessionNotFound()
	}

	isValid := service.isChosenHeroesValid(session.AllowedHeroList, chosen)
	if !isValid {
		return nil, exceptions.HeroPickupError()
	}

	heroList := make([]*models.Hero, 0)
	for _, heroEnum := range chosen {
		baseHero := service.HeroFactory.Create(heroEnum)
		baseHeroInfo := baseHero.GetInfo()
		hero := &models.Hero{
			Health:         baseHeroInfo.BaseStats.MaxHealth,
			LastMoveTurn:   -2,
			LastAttackTurn: -2,
			BaseHero:       heroEnum,
			PlayerId:       playerId,
			SessionId:      session.Id,
		}
		_, err := service.HeroRepository.SaveHero(tx, hero)
		if err != nil {
			return nil, err
		}
		heroList = append(heroList, hero)
	}

	err = service.PlayerService.SetPlayerHeroes(tx, sessionId, playerId, chosen)
	if err != nil {
		return nil, err
	}

	return heroList, nil
}

func (service *HeroServiceImpl) InitHeroPosition(
	tx databases.BadgerTx,
	heroList1 []*models.Hero,
	spawnPoints1 []physics.Point,
	heroList2 []*models.Hero,
	spawnPoints2 []physics.Point,
) error {
	log.Println("Setting up hero spawns...")
	log.Println(heroList1)
	log.Println(spawnPoints1)
	log.Println(heroList2)
	log.Println(spawnPoints2)

	for i, spawnPoint := range spawnPoints1 {
		if i < len(heroList1) {
			heroList1[i].Position = spawnPoint
			_, err := service.HeroRepository.SaveHero(tx, heroList1[i])
			if err != nil {
				return err
			}
			_, err = service.HeroRepository.SaveInitialHero(tx, heroList1[i])
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
			_, err := service.HeroRepository.SaveHero(tx, heroList2[i])
			if err != nil {
				return err
			}
			_, err = service.HeroRepository.SaveInitialHero(tx, heroList2[i])
			if err != nil {
				return err
			}
		} else {
			break
		}
	}

	log.Println("After setting up...")
	log.Println(heroList1)
	log.Println(spawnPoints1)
	log.Println(heroList2)
	log.Println(spawnPoints2)

	return nil
}

func (service *HeroServiceImpl) GetInitialState(
	tx databases.BadgerTx,
	heroList1 []*models.Hero,
	heroList2 []*models.Hero,
) ([]*models.Hero, []*models.Hero, error) {
	initList1 := make([]*models.Hero, 0)
	initList2 := make([]*models.Hero, 0)
	for _, hero := range heroList1 {
		obj, err := service.HeroRepository.GetInitialHero(tx, hero.SessionId, hero.PlayerId, hero.BaseHero)
		if err != nil {
			return nil, nil, err
		}
		initList1 = append(initList1, obj)
	}
	for _, hero := range heroList2 {
		obj, err := service.HeroRepository.GetInitialHero(tx, hero.SessionId, hero.PlayerId, hero.BaseHero)
		if err != nil {
			return nil, nil, err
		}
		initList2 = append(initList2, obj)
	}
	return initList1, initList2, nil
}

func (service *HeroServiceImpl) SetHeroes(tx databases.BadgerTx, heroList []*models.Hero) error {
	_, err := service.HeroRepository.SaveHeroes(tx, heroList)
	if err != nil {
		return err
	}
	return nil
}

// Checks the validity of number of heroes, availability, and duplication.
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

func (service *HeroServiceImpl) SetSessionService(sessionService SessionService) {
	service.SessionService = sessionService
}

func (service *HeroServiceImpl) SetPlayerService(playerService PlayerService) {
	service.PlayerService = playerService
}

func NewHeroService(
	heroRepository repositories.HeroRepository,
	sessionService SessionService,
	playerService PlayerService,
	heroFactory heroes.BaseHeroFactory,
	transactionManager databases.TransactionManager,
) HeroService {
	heroService := &HeroServiceImpl{
		HeroRepository:     heroRepository,
		SessionService:     sessionService,
		PlayerService:      playerService,
		HeroFactory:        heroFactory,
		TransactionManager: transactionManager,
	}
	return heroService
}
