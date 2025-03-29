package services

import (
	"errors"

	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/events"
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
	GetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string) ([]*models.Hero, error)
	GetPlayerHero(tx databases.BadgerTx, sessionId string, playerId string, baseHero heroes.BaseHeroEnum) (*models.Hero, error)
	GetAvailableHeroes() []heroes.BaseHeroEnum
	CreateHeroesTx(tx databases.BadgerTx, sessionId string, playerId string, chosen []heroes.BaseHeroEnum) error
	CreateHeroes(sessionId string, playerId string, chosen []heroes.BaseHeroEnum) error
	InitHeroPosition(heroList1 []*models.Hero, spawnPoints1 []physics.Point, heroList2 []*models.Hero, spawnPoints2 []physics.Point) error
	InitHeroPositionTx(tx databases.BadgerTx, heroList1 []*models.Hero, spawnPoints1 []physics.Point, heroList2 []*models.Hero, spawnPoints2 []physics.Point) error
	GetInitialState(tx databases.BadgerTx, heroList1 []*models.Hero, heroList2 []*models.Hero) ([]*models.Hero, []*models.Hero, error)

	ApplyDamage(tx databases.BadgerTx, currentTurn int, srcHero *models.Hero, trgHero *models.Hero, damage int) error
	MoveHero(tx databases.BadgerTx, currentTurn int, srcHero *models.Hero, position physics.Point) error

	SetSessionService(sessionService SessionService)
	SetPlayerService(playerService PlayerService)
}

type HeroServiceImpl struct {
	HeroRepository repositories.HeroRepository
	SessionService SessionService
	PlayerService  PlayerService

	HeroFactory        heroes.BaseHeroFactory
	TransactionManager databases.TransactionManager
	EventManager       events.EventManager
}

type AttackEvent struct {
	Tx          databases.BadgerTx
	CurrentTurn int
	SrcHero     *models.Hero
	DstHero     *models.Hero
	Damage      int
}

type MoveEvent struct {
	Tx          databases.BadgerTx
	CurrentTurn int
	SrcHero     *models.Hero
	Position    physics.Point
}

func (service *HeroServiceImpl) GetPlayerHeroes(tx databases.BadgerTx, sessionId string, playerId string) ([]*models.Hero, error) {
	player := service.PlayerService.GetPlayer(tx, sessionId, playerId)
	if player == nil {
		return nil, errors.New("invalid player")
	}

	return service.HeroRepository.GetPlayerHeroes(nil, sessionId, playerId, player.HeroBases)
}

func (service *HeroServiceImpl) GetPlayerHero(tx databases.BadgerTx, sessionId string, playerId string, baseHero heroes.BaseHeroEnum) (*models.Hero, error) {
	return service.HeroRepository.GetHeroBySessionId(nil, repositories.HeroKey{
		SessionId: sessionId,
		PlayerId:  playerId,
		BaseHero:  baseHero,
	})
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

func (service *HeroServiceImpl) onMove(data interface{}) error {
	moveData, ok := data.(*MoveEvent)
	if !ok {
		return errors.New("invalid data")
	}

	err := service.MoveHero(moveData.Tx, moveData.CurrentTurn, moveData.SrcHero, moveData.Position)
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

func (service *HeroServiceImpl) onDamage(data interface{}) error {
	attackData, ok := data.(*AttackEvent)
	if !ok {
		return errors.New("invalid data")
	}

	err := service.ApplyDamage(attackData.Tx, attackData.CurrentTurn, attackData.SrcHero, attackData.DstHero, attackData.Damage)
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
	// TODO: check whether not assigned returns error too or nil.. if error, which error (repo)
	session, err := service.SessionService.GetSessionById(tx, sessionId)
	if err != nil {
		return err
	}
	if session == nil {
		return exceptions.SessionNotFound()
	}

	isValid := service.isChosenHeroesValid(session.AllowedHeroList, chosen)
	if !isValid {
		return exceptions.HeroPickupError()
	}

	for _, heroEnum := range chosen {
		baseHero := service.HeroFactory.Create(heroEnum)
		baseHeroInfo := baseHero.GetInfo()
		_, err := service.HeroRepository.SaveHero(tx, &models.Hero{
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

	err = service.PlayerService.SetPlayerHeroes(tx, sessionId, playerId, chosen)
	if err != nil {
		return err
	}

	return nil
}

func (service *HeroServiceImpl) CreateHeroes(sessionId string, playerId string, chosen []heroes.BaseHeroEnum) error {
	tx := service.TransactionManager.NewReadWriteTransaction()
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

	return nil
}

func (service *HeroServiceImpl) InitHeroPosition(
	heroList1 []*models.Hero,
	spawnPoints1 []physics.Point,
	heroList2 []*models.Hero,
	spawnPoints2 []physics.Point,
) error {
	tx := service.TransactionManager.NewReadWriteTransaction()
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
	// heroList := heroList1
	// heroList = append(heroList, heroList2...)

	// for i, hero := range heroList {
	// 	if i < len(heroList1) {
	// 		_, err := service.HeroRepository.GetInitialHero(tx, hero.SessionId, hero.PlayerId, hero.BaseHero)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		_, err = service.HeroRepository.SaveHero(tx, hero)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	} else {
	// 		break
	// 	}
	// }

	// return nil
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
	eventManager events.EventManager,
) HeroService {
	heroService := &HeroServiceImpl{
		HeroRepository:     heroRepository,
		SessionService:     sessionService,
		PlayerService:      playerService,
		HeroFactory:        heroFactory,
		TransactionManager: transactionManager,
		EventManager:       eventManager,
	}
	eventManager.On("MOVE", heroService.onMove)
	eventManager.On("DAMAGE", heroService.onDamage)
	return heroService
}
