package services

import (
	"errors"
	"log"
	"time"

	"pixeltactics.com/match/src/core/states"
	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/heroes"
	"pixeltactics.com/match/src/models"
	"pixeltactics.com/match/src/repositories"
	"pixeltactics.com/match/src/utils/physics"
)

const (
	PreparationTime = 10 * time.Second
	PlayerTurnTime  = 30 * time.Second
)

type SessionService interface {
	GetSessionById(sessionId string) *models.Session
}

type SessionServiceImpl struct {
	mapService        MapService
	heroService       HeroService
	playerService     PlayerService
	sessionRepository repositories.SessionRepositoryV2

	stateFactory states.SessionStateFactory
}

func (service *SessionServiceImpl) GetSessionById(sessionId string) *models.Session {
	return service.sessionRepository.GetSessionById(sessionId)
}

func (service *SessionServiceImpl) GetSessionByPlayerId(playerId string) *models.Session {
	return service.sessionRepository.GetSessionByPlayerId(playerId)
}

func (service *SessionServiceImpl) CreateSession(playerId string, opponentId string) (*models.Session, error) {
	err := service.checkPlayerSession(playerId)
	if err != nil {
		return nil, err
	}

	isStart, err := service.checkOpponentSession(playerId, opponentId)
	if err != nil {
		return nil, err
	}

	// Session object is already created (by opponent)
	if isStart {
		oppSession := service.sessionRepository.GetSessionByPlayerId(opponentId)
		if oppSession == nil {
			log.Fatalln("session found before but now not")
			return nil, errors.New("server cannot get opponent session")
		}
		service.runSession(oppSession)
		return oppSession, nil
	}

	// Session object is not created, so create it
	availHeroes := service.heroService.GetAvailableHeroes()
	session, err := service.sessionRepository.CreateSession(repositories.CreateSessionParams{
		PlayerId1:         playerId,
		PlayerId2:         opponentId,
		AvailableHeroList: availHeroes,
	})
	if err != nil {
		return nil, err
	}

	err = service.playerService.CreatePlayerForSession(session.Id, playerId, opponentId)
	if err != nil {
		service.sessionRepository.DeleteSession(session.Id)
		return nil, err
	}

	_, err = service.mapService.GenerateMap(session.Id)
	if err != nil {
		service.sessionRepository.DeleteSession(session.Id)
		return nil, err
	}

	return session, nil
}

func (service *SessionServiceImpl) runSession(session *models.Session) {
	preparationDeadline := time.Now().Add(PreparationTime)
	sessionState := service.stateFactory.Create(session)
	err := sessionState.Start(preparationDeadline)
	if err != nil {
		panic("PANIC: invalid session state")
	}
	service.sessionRepository.UpdateSession(repositories.UpdateSessionParams{
		SessionId: session.Id,
		State:     session.State,
	})
	time.AfterFunc(time.Until(preparationDeadline), func() {
		service.checkForExpire(session, session.State.Id)
	})
}

func (service *SessionServiceImpl) PreparePlayer(playerId string, chosenHeroes []heroes.BaseHeroEnum) error {
	session := service.sessionRepository.GetSessionByPlayerId(playerId)
	if session == nil || session.State.Type != models.SessionStatePreparation {
		return exceptions.ActionNotAllowed()
	}

	if time.Now().After(session.State.Deadline) {
		return exceptions.ExceededDeadlineError()
	}

	service.heroService.CreateHeroes(session.Id, playerId, chosenHeroes)
	return nil
}

func (service *SessionServiceImpl) StartBattle(session *models.Session) error {
	if session.State.Type != models.SessionStatePreparation {
		return exceptions.ActionNotAllowed()
	}

	if time.Now().After(session.State.Deadline) {
		return exceptions.ExceededDeadlineError()
	}

	player1 := service.playerService.GetPlayer(session.PlayerIds[0], session.Id)
	player2 := service.playerService.GetPlayer(session.PlayerIds[1], session.Id)

	// TODO: check if empty = error
	heroList1, err1 := service.heroService.GetPlayerHeroes(session.Id, player1.Id)
	heroList2, err2 := service.heroService.GetPlayerHeroes(session.Id, player2.Id)
	if err1 != nil || err2 != nil {
		return exceptions.HeroPickupError()
	}

	sessionMap, err := service.mapService.GetSessionMap(session.Id)
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

	err = service.heroService.InitHeroPosition(heroList1, spawnPoints1, heroList2, spawnPoints2)
	if err != nil {
		return err
	}

	sessionState := service.stateFactory.Create(session)
	sessionState.Start(time.Now().Add(PlayerTurnTime))
	service.sessionRepository.UpdateSession(repositories.UpdateSessionParams{
		SessionId: session.Id,
		State:     session.State,
		WinnerId:  session.WinnerId,
	})

	return nil
}

func (service *SessionServiceImpl) checkForExpire(session *models.Session, lastStateId string) error {
	if session.State.Id != lastStateId {
		return nil
	}

	sessionState := service.stateFactory.Create(session)
	err := sessionState.End(nil)
	if err != nil {
		return err
	}
	service.sessionRepository.UpdateSession(repositories.UpdateSessionParams{
		SessionId: session.Id,
		State:     session.State,
	})
	return nil
}

func (service *SessionServiceImpl) checkPlayerSession(playerId string) error {
	plrSession := service.sessionRepository.GetSessionByPlayerId(playerId)
	if plrSession != nil && plrSession.IsRunning() {
		return errors.New("player is in running session")
	}
	return nil
}

func (service *SessionServiceImpl) checkOpponentSession(playerId string, opponentId string) (bool, error) {
	oppSession := service.sessionRepository.GetSessionByPlayerId(opponentId)
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
