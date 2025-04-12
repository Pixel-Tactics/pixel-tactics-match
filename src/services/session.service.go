package services

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
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
	// When no key found (or empty) for session, nil session will be returned instead of error.
	GetSessionById(tx databases.BadgerTx, sessionId string) (*models.Session, error)

	// When no key found (or empty) for session, nil session will be returned instead of error.
	GetSessionByPlayerId(tx databases.BadgerTx, playerId string) (*models.Session, error)

	// Creates session object for player and opponent. If it was empty, player invites opponent.
	// But if the opponent already invited player, it will start the match by going into preparation state.
	CreateSession(playerId string, opponentId string) (*models.Session, error)

	PreparePlayer(playerId string, chosenHeroes []heroes.BaseHeroEnum) (bool, error)
	CompileSession(sessionId string) (map[string]interface{}, error)
}

type SessionServiceImpl struct {
	MapService        MapService
	HeroService       HeroService
	PlayerService     PlayerService
	SessionRepository repositories.SessionRepositoryV2
	LogRepository     repositories.SessionLogRepository

	StateFactory       states.SessionStateFactory
	TransactionManager databases.TransactionManager
}

// When no key found (or empty) for session, nil session will be returned instead of error.
func (service *SessionServiceImpl) GetSessionById(tx databases.BadgerTx, sessionId string) (*models.Session, error) {
	return service.SessionRepository.GetSessionById(tx, sessionId)
}

// When no key found (or empty) for session, nil session will be returned instead of error.
func (service *SessionServiceImpl) GetSessionByPlayerId(tx databases.BadgerTx, playerId string) (*models.Session, error) {
	return service.SessionRepository.GetSessionByPlayerId(tx, playerId)
}

// Creates session object for player and opponent. If it was empty, player invites opponent.
// But if the opponent already invited player, it will start the match by going into preparation state.
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
		log.Println("STARTING...")
		oppSession, err := service.SessionRepository.GetSessionByPlayerId(tx, opponentId)
		if err != nil {
			log.Fatalln(err)
			return nil, err
		}
		if oppSession == nil {
			log.Fatalln("session found before but now not")
			return nil, errors.New("server cannot get opponent session")
		}
		err = service.runSession(tx, oppSession)
		if err != nil {
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
		return oppSession, nil
	}

	// Session object is not created, so create it
	availHeroes := service.HeroService.GetAvailableHeroes()
	sessionId := uuid.New().String()
	session := &models.Session{
		Id: sessionId,
		State: models.State{
			Id:   uuid.New().String(),
			Type: models.SessionStateMatchMaking,
		},
		AllowedHeroList: availHeroes,
		PlayerIds: []string{
			playerId,
			opponentId,
		},
	}
	_, err = service.SessionRepository.SaveSession(tx, session)
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

// Start the preparation state
func (service *SessionServiceImpl) runSession(tx databases.BadgerTx, session *models.Session) error {
	preparationDeadline := time.Now().Add(PreparationTime)
	sessionState := service.StateFactory.Create(session)
	err := sessionState.Start(preparationDeadline)
	if err != nil {
		panic("PANIC: invalid session state")
	}

	stateLog, err := convert_utils.ObjectToMap(session.State)
	if err != nil {
		return err
	}

	_, err = service.SessionRepository.SaveSession(tx, session)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = service.LogRepository.AppendLog(tx, session.Id, &models.SessionLog{
		Type:      "STATE_CHANGE",
		SessionId: session.Id,
		Data:      stateLog,
	})
	if err != nil {
		return err
	}

	log.Println(session)
	time.AfterFunc(time.Until(preparationDeadline), func() {
		service.checkForExpire(tx, session, session.State.Id)
	})
	return nil
}

// Chooses heroes for player. It returns boolean that represents whether the preparation ends (other player have chosen their heroes too) or not.
func (service *SessionServiceImpl) PreparePlayer(playerId string, chosenHeroes []heroes.BaseHeroEnum) (bool, error) {
	tx := service.TransactionManager.NewReadWriteTransaction()
	defer tx.Discard()

	session, err := service.SessionRepository.GetSessionByPlayerId(tx, playerId)
	if err != nil {
		return false, err
	}
	if session == nil || session.State.Type != models.SessionStatePreparation {
		return false, exceptions.ActionNotAllowed()
	}

	if time.Now().After(session.State.Deadline) {
		return false, exceptions.ExceededDeadlineError()
	}

	heroList, err := service.HeroService.CreateHeroes(tx, session.Id, playerId, chosenHeroes)
	if err != nil {
		return false, err
	}

	err = service.startBattle(tx, playerId, heroList, session)
	log.Println(err)

	var otherPlayerNotPicked = false
	if err != nil {
		otherPlayerNotPicked = err.Error() == exceptions.HeroPickupError().Error()
	}

	log.Println(otherPlayerNotPicked)

	if err == nil || otherPlayerNotPicked {
		err = tx.Commit()
		if err != nil {
			log.Println("Error when commiting player preparation")
			return false, err
		}
		log.Println("Commited Preparation..")
		return !otherPlayerNotPicked, nil
	}
	return false, err
}

func (service *SessionServiceImpl) startBattle(tx databases.BadgerTx, playerId string, playerHeroList []*models.Hero, session *models.Session) error {
	if session.State.Type != models.SessionStatePreparation {
		return exceptions.ActionNotAllowed()
	}

	if time.Now().After(session.State.Deadline) {
		return exceptions.ExceededDeadlineError()
	}

	player1, err := service.PlayerService.GetPlayer(tx, session.PlayerIds[0], session.Id)
	if err != nil {
		return err
	}
	if player1 == nil {
		return errors.New("invalid player 1")
	}
	player2, err := service.PlayerService.GetPlayer(tx, session.PlayerIds[1], session.Id)
	if err != nil {
		return err
	}
	if player2 == nil {
		return errors.New("invalid player 1")
	}

	var heroList1 []*models.Hero
	var heroList2 []*models.Hero
	if playerId == session.PlayerIds[0] {
		heroList1 = playerHeroList
		heroList2, err := service.HeroService.GetPlayerHeroes(tx, session.Id, player2.Id)
		if err != nil || len(heroList2) == 0 {
			log.Println("Cannot start match yet (" + err.Error() + ")..")
			return exceptions.HeroPickupError()
		}
	} else {
		heroList1, err := service.HeroService.GetPlayerHeroes(tx, session.Id, player1.Id)
		heroList2 = playerHeroList
		if err != nil || len(heroList1) == 0 {
			log.Println("Cannot start match yet (" + err.Error() + ")..")
			return exceptions.HeroPickupError()
		}
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

	stateLog, err := convert_utils.ObjectToMap(session.State)
	if err != nil {
		return err
	}

	_, err = service.SessionRepository.SaveSession(tx, session)
	if err != nil {
		return err
	}
	_, err = service.LogRepository.AppendLog(tx, session.Id, &models.SessionLog{
		Type:      "STATE_CHANGE",
		SessionId: session.Id,
		Data:      stateLog,
	})
	if err != nil {
		return err
	}

	_, err = service.SessionRepository.SaveInitialSession(tx, session)
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

	stateLog, err := convert_utils.ObjectToMap(session.State)
	if err != nil {
		return err
	}

	_, err = service.SessionRepository.SaveSession(tx, session)
	if err != nil {
		return err
	}
	_, err = service.LogRepository.AppendLog(tx, session.Id, &models.SessionLog{
		Type:      "STATE_CHANGE",
		SessionId: session.Id,
		Data:      stateLog,
	})
	if err != nil {
		return err
	}
	return nil
}

func (service *SessionServiceImpl) checkPlayerSession(tx databases.BadgerTx, playerId string) error {
	plrSession, err := service.SessionRepository.GetSessionByPlayerId(tx, playerId)
	if err != nil {
		return err
	}
	if plrSession != nil && plrSession.IsRunning() {
		return errors.New("player is in running session")
	}
	return nil
}

func (service *SessionServiceImpl) checkOpponentSession(tx databases.BadgerTx, playerId string, opponentId string) (bool, error) {
	oppSession, err := service.SessionRepository.GetSessionByPlayerId(tx, opponentId)
	if err != nil {
		return false, err
	}
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
	session, err := service.GetSessionById(nil, sessionId)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}
	state, err := convert_utils.ObjectToMap(session.State)
	if err != nil {
		return nil, err
	}

	matchMap, err := service.MapService.GetSessionMap(nil, sessionId)
	if err != nil {
		return nil, err
	}
	heroList1, err := service.HeroService.GetPlayerHeroes(nil, sessionId, session.PlayerIds[0])
	if err != nil && err == ErrEmptyHero {
		heroList1 = make([]*models.Hero, 0)
	} else if err != nil {
		return nil, err
	}
	heroList2, err := service.HeroService.GetPlayerHeroes(nil, sessionId, session.PlayerIds[1])
	if err != nil && err == ErrEmptyHero {
		heroList2 = make([]*models.Hero, 0)
	} else if err != nil {
		return nil, err
	}

	logs, err := service.LogRepository.GetSessionLogs(nil, sessionId)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id": session.Id,
		"player1": map[string]interface{}{
			"id":       session.PlayerIds[0],
			"heroList": heroList1,
		},
		"player2": map[string]interface{}{
			"id":       session.PlayerIds[1],
			"heroList": heroList2,
		},
		"state":             state,
		"availableHeroList": session.AllowedHeroList,
		"matchMap":          matchMap,
		"logs":              logs,
	}, nil
}

func NewSessionService(
	mapService MapService,
	heroService HeroService,
	playerService PlayerService,
	sessionRepository repositories.SessionRepositoryV2,
	stateFactory states.SessionStateFactory,
	transactionManager databases.TransactionManager,
	logRepository repositories.SessionLogRepository,
) SessionService {
	return &SessionServiceImpl{
		MapService:         mapService,
		HeroService:        heroService,
		PlayerService:      playerService,
		SessionRepository:  sessionRepository,
		StateFactory:       stateFactory,
		TransactionManager: transactionManager,
		LogRepository:      logRepository,
	}
}
