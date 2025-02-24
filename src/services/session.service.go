package services

import (
	"errors"
	"log"
	"time"

	"pixeltactics.com/match/src/core/states"
	"pixeltactics.com/match/src/databases"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
	convert_utils "pixeltactics.com/match/src/utils/convert"
	"pixeltactics.com/match/src/utils/physics"
)

const (
	PreparationTime = 10 * time.Second
	PlayerTurnTime  = 30 * time.Second
)

type SessionService interface {
	GetSessionById(tx databases.BadgerTx, sessionId string) *models.Session
	GetSessionByPlayerId(playerId string) *models.Session
	CreateSession(playerId string, opponentId string) (*models.Session, error)
	CompileSession(sessionId string) (map[string]interface{}, error)
}

type SessionServiceImpl struct {
	MapService        MapService
	HeroService       HeroService
	PlayerService     PlayerService
	SessionRepository repositories.SessionRepositoryV2

	StateFactory       states.SessionStateFactory
	TransactionManager databases.TransactionManager
}

func (service *SessionServiceImpl) GetSessionById(tx databases.BadgerTx, sessionId string) *models.Session {
	return service.SessionRepository.GetSessionById(tx, sessionId)
}

func (service *SessionServiceImpl) GetSessionByPlayerIdTx(tx databases.BadgerTx, playerId string) *models.Session {
	return service.SessionRepository.GetSessionByPlayerId(tx, playerId)
}

func (service *SessionServiceImpl) GetSessionByPlayerId(playerId string) *models.Session {
	return service.SessionRepository.GetSessionByPlayerId(nil, playerId)
}

func (service *SessionServiceImpl) CreateSession(playerId string, opponentId string) (*models.Session, error) {
	tx := service.TransactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	err := service.checkPlayerSession(tx, playerId)
	if err != nil {
		return nil, err
	}

	isStart, err := service.checkOpponentSession(tx, playerId, opponentId)
	if err != nil {
		return nil, err
	}

	// Session object is already created (by opponent)
	if isStart {
		oppSession := service.SessionRepository.GetSessionByPlayerId(tx, opponentId)
		if oppSession == nil {
			log.Fatalln("session found before but now not")
			return nil, errors.New("server cannot get opponent session")
		}
		service.runSession(tx, oppSession)
		return oppSession, nil
	}

	// Session object is not created, so create it
	availHeroes := service.HeroService.GetAvailableHeroes()
	session, err := service.SessionRepository.CreateSession(tx, repositories.CreateSessionParams{
		PlayerId1:         playerId,
		PlayerId2:         opponentId,
		AvailableHeroList: availHeroes,
	})
	if err != nil {
		return nil, err
	}

	err = service.PlayerService.CreatePlayerForSession(tx, session.Id, playerId, opponentId)
	if err != nil {
		service.SessionRepository.DeleteSession(tx, session.Id)
		return nil, err
	}

	_, err = service.MapService.GenerateMap(tx, session.Id)
	if err != nil {
		service.SessionRepository.DeleteSession(tx, session.Id)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (service *SessionServiceImpl) runSession(tx databases.BadgerTx, session *models.Session) {
	preparationDeadline := time.Now().Add(PreparationTime)
	sessionState := service.StateFactory.Create(session)
	err := sessionState.Start(preparationDeadline)
	if err != nil {
		panic("PANIC: invalid session state")
	}
	service.SessionRepository.UpdateSession(tx, repositories.UpdateSessionParams{
		SessionId: session.Id,
		State:     session.State,
	})
	time.AfterFunc(time.Until(preparationDeadline), func() {
		service.checkForExpire(tx, session, session.State.Id)
	})
}

func (service *SessionServiceImpl) PreparePlayer(playerId string, chosenHeroes []heroes.BaseHeroEnum) (bool, error) {
	tx := service.TransactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	session := service.SessionRepository.GetSessionByPlayerId(tx, playerId)
	if session == nil || session.State.Type != models.SessionStatePreparation {
		return false, exceptions.ActionNotAllowed()
	}

	if time.Now().After(session.State.Deadline) {
		return false, exceptions.ExceededDeadlineError()
	}

	err := service.HeroService.CreateHeroesTx(tx, session.Id, playerId, chosenHeroes)
	if err != nil {
		return false, err
	}

	err = service.StartBattle(tx, session)

	var otherPlayerNotPicked = false
	if err != nil {
		otherPlayerNotPicked = err.Error() == exceptions.HeroPickupError().Error()
	}

	if err == nil || otherPlayerNotPicked {
		err = tx.Commit()
		if err != nil {
			return false, err
		}
		return !otherPlayerNotPicked, nil
	}
	return false, err
}

func (service *SessionServiceImpl) StartBattle(tx databases.BadgerTx, session *models.Session) error {
	if session.State.Type != models.SessionStatePreparation {
		return exceptions.ActionNotAllowed()
	}

	if time.Now().After(session.State.Deadline) {
		return exceptions.ExceededDeadlineError()
	}

	player1 := service.PlayerService.GetPlayer(tx, session.PlayerIds[0], session.Id)
	player2 := service.PlayerService.GetPlayer(tx, session.PlayerIds[1], session.Id)

	// TODO: check if empty = error
	heroList1, err1 := service.HeroService.GetPlayerHeroes(tx, session.Id, player1.Id)
	heroList2, err2 := service.HeroService.GetPlayerHeroes(tx, session.Id, player2.Id)
	if err1 != nil || err2 != nil {
		return exceptions.HeroPickupError()
	}

	sessionMap, err := service.MapService.GetSessionMap(tx, session.Id)
	if err != nil {
		return err
	}

	spawnPoints1 := make([]physics.Point, 0)
	spawnPoints2 := make([]physics.Point, 0)

	for i := range sessionMap.Structure {
		for j := range sessionMap.Structure[0] {
			if sessionMap.Structure[i][j] == 3 {
				spawnPoints1 = append(spawnPoints1, physics.Point{X: j, Y: i})
			} else if sessionMap.Structure[i][j] == 4 {
				spawnPoints2 = append(spawnPoints2, physics.Point{X: j, Y: i})
			}
		}
	}

	err = service.HeroService.InitHeroPositionTx(tx, heroList1, spawnPoints1, heroList2, spawnPoints2)
	if err != nil {
		return err
	}

	sessionState := service.StateFactory.Create(session)
	err = sessionState.Start(time.Now().Add(PlayerTurnTime))
	if err != nil {
		return err
	}

	_, err = service.SessionRepository.UpdateSession(tx, repositories.UpdateSessionParams{
		SessionId: session.Id,
		State:     session.State,
		WinnerId:  session.WinnerId,
	})
	if err != nil {
		return err
	}

	return nil
}

func (service *SessionServiceImpl) checkForExpire(tx databases.BadgerTx, session *models.Session, lastStateId string) error {
	if session.State.Id != lastStateId {
		return nil
	}

	sessionState := service.StateFactory.Create(session)
	err := sessionState.End(nil)
	if err != nil {
		return err
	}
	service.SessionRepository.UpdateSession(tx, repositories.UpdateSessionParams{
		SessionId: session.Id,
		State:     session.State,
	})
	return nil
}

func (service *SessionServiceImpl) checkPlayerSession(tx databases.BadgerTx, playerId string) error {
	plrSession := service.SessionRepository.GetSessionByPlayerId(tx, playerId)
	if plrSession != nil && plrSession.IsRunning() {
		return errors.New("player is in running session")
	}
	return nil
}

func (service *SessionServiceImpl) checkOpponentSession(tx databases.BadgerTx, playerId string, opponentId string) (bool, error) {
	oppSession := service.SessionRepository.GetSessionByPlayerId(tx, opponentId)
	if oppSession != nil {
		if oppSession.IsRunning() {
			return false, errors.New("opponent is in running session")
		} else {
			oppSessionOtherId, _ := oppSession.GetOtherPlayerId(opponentId)
			return oppSessionOtherId == playerId, nil // start match if opponent of opponent is player
		}
	}
	return false, nil
}

func (service *SessionServiceImpl) CompileSession(sessionId string) (map[string]interface{}, error) {
	// actionLogData := []map[string]interface{}{}
	// for i, actionLog := range session.actionLog {
	// 	actionData := actionLog.GetData()
	// 	actionData["order"] = i
	// 	actionLogData = append(actionLogData, actionData)
	// }
	session := service.GetSessionById(nil, sessionId)
	state, err := convert_utils.ObjectToMap(session.State)
	if err != nil {
		return nil, err
	}

	matchMap, err := service.MapService.GetSessionMap(nil, sessionId)
	if err != nil {
		return nil, err
	}
	// heroList1, err := service.HeroService.GetPlayerHeroes(nil, sessionId, session.PlayerIds[0])
	// if err != nil {
	// 	return nil, err
	// }
	// compiledHero1 := make([]map[string]interface{}, 0)
	// for _, hero := range heroList1 {
	// 	compiledHero1 = append(compiledHero1, map[string]interface{}{

	// 	})
	// }
	// heroList2, err := service.HeroService.GetPlayerHeroes(nil, sessionId, session.PlayerIds[1])
	// if err != nil {
	// 	return nil, err
	// }
	return map[string]interface{}{
		"id": session.Id,
		// "player1": map[string]interface{}{
		// 	"id":       session.PlayerIds[0],
		// 	"heroList": heroList1,
		// },
		// "player2": map[string]interface{}{
		// 	"id":       session.PlayerIds[1],
		// 	"heroList": heroList2,
		// },
		"state":             state,
		"availableHeroList": session.AllowedHeroList,
		"matchMap":          matchMap,
		"actionLog":         make([]map[string]interface{}, 0),
	}, nil
}

// func (p *Player) GetData() map[string]interface{} {
// 	var heroListData = []map[string]interface{}{}
// 	for _, hero := range p.HeroList {
// 		heroListData = append(heroListData, hero.GetData())
// 	}
// 	return map[string]interface{}{
// 		"id":       p.Id,
// 		"heroList": heroListData,
// 	}
// }

func NewSessionService(
	mapService MapService,
	heroService HeroService,
	playerService PlayerService,
	sessionRepository repositories.SessionRepositoryV2,
	stateFactory states.SessionStateFactory,
	transactionManager databases.TransactionManager,
) SessionService {
	return &SessionServiceImpl{
		MapService:         mapService,
		HeroService:        heroService,
		PlayerService:      playerService,
		SessionRepository:  sessionRepository,
		StateFactory:       stateFactory,
		TransactionManager: transactionManager,
	}
}
