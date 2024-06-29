package repositories

import (
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
	"pixeltactics.com/match/src/data_structures"
	"pixeltactics.com/match/src/matches"
	matches_heroes_templates "pixeltactics.com/match/src/matches/heroes/templates"
	matches_maps "pixeltactics.com/match/src/matches/maps"
)

type SessionRepository struct {
	sessions      *data_structures.SyncMap[string, *matches.Session]
	playerSession *data_structures.SyncMap[string, *matches.Session]
}

func (repo *SessionRepository) GetSessionById(id string) *matches.Session {
	value, ok := repo.sessions.Load(id)
	if !ok {
		return nil
	}
	return value
}

func (repo *SessionRepository) GetSessionByPlayerId(playerId string) *matches.Session {
	value, ok := repo.playerSession.Load(playerId)
	if !ok {
		return nil
	}
	return value
}

func (repo *SessionRepository) CreateSession(playerId string, opponentId string) (*matches.Session, error) {
	err := repo.checkPlayerSession(playerId, opponentId)
	if err != nil {
		return nil, err
	}

	isStart, err := repo.checkOpponentSession(playerId, opponentId)
	if err != nil {
		return nil, err
	}

	if isStart {
		opponentSession, isOpponentInSession := repo.playerSession.Load(opponentId)
		if !isOpponentInSession {
			log.Fatalln("session found before but now not")
			return nil, errors.New("server cannot get opponent session")
		}
		repo.playerSession.Store(playerId, opponentSession)
		opponentSession.RunSessionSync()
		return opponentSession, nil
	}

	newSessionId := uuid.New().String()
	_, ok := repo.sessions.Load(newSessionId)
	if ok {
		return nil, errors.New("duplicate session id")
	}

	matchMap, err := matches_maps.GenerateMap()
	if err != nil {
		return nil, err
	}

	availableHeroList, err := matches_heroes_templates.GetAvailableHeroes()
	if err != nil {
		return nil, err
	}

	newSession := matches.NewSession(newSessionId, playerId, opponentId, matchMap, availableHeroList)

	repo.sessions.Store(newSessionId, newSession)
	repo.playerSession.Store(playerId, newSession)
	return newSession, nil
}

func (repo *SessionRepository) DeleteSession(sessionId string) {
	session := repo.GetSessionById(sessionId)
	if session == nil {
		return
	}

	player1Id, player2Id := session.GetPlayersSync()
	session1, ok1 := repo.playerSession.Load(player1Id)
	if ok1 && session1 == session {
		repo.playerSession.Delete(player1Id)
	}
	session2, ok2 := repo.playerSession.Load(player2Id)
	if ok2 && session2 == session {
		repo.playerSession.Delete(player2Id)
	}

	repo.sessions.Delete(sessionId)
}

func (repo *SessionRepository) checkPlayerSession(playerId string, opponentId string) error {
	playerSession, isPlayerInSession := repo.playerSession.Load(playerId)
	if isPlayerInSession {
		sessionOpponent, _ := playerSession.GetOpponentPlayerSync(playerId)
		isRunning := playerSession.GetRunningSync()

		if !isRunning {
			repo.DeleteSession(playerSession.GetIdSync())
			return nil
		} else {
			if sessionOpponent.Id != opponentId {
				return errors.New("player is already on a session with another player")
			} else {
				return errors.New("session is already created")
			}
		}
	}
	return nil
}

func (repo *SessionRepository) checkOpponentSession(playerId string, opponentId string) (bool, error) {
	opponentSession, isOpponentInSession := repo.playerSession.Load(opponentId)
	if isOpponentInSession {
		sessionOpponent, _ := opponentSession.GetOpponentPlayerSync(opponentId)
		isRunning := opponentSession.GetRunningSync()
		isEnded := opponentSession.GetEndedSync()

		if isEnded {
			repo.DeleteSession(opponentSession.GetIdSync())
			return false, nil
		} else if isRunning {
			if sessionOpponent.Id != playerId {
				return false, errors.New("opponent is already on a session with another player")
			} else {
				return false, errors.New("session is already created")
			}
		} else {
			return sessionOpponent.Id == playerId, nil
		}
	}
	return false, nil
}

var sessionRepository *SessionRepository = nil
var onceSession sync.Once

func GetSessionRepository() *SessionRepository {
	onceSession.Do(func() {
		sessionRepository = &SessionRepository{
			sessions:      data_structures.NewSyncMap[string, *matches.Session](),
			playerSession: data_structures.NewSyncMap[string, *matches.Session](),
		}
	})
	return sessionRepository
}
