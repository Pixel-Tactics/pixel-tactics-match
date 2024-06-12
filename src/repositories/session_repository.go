package repositories

import (
	"errors"

	"github.com/google/uuid"
	"pixeltactics.com/match/src/data_structures"
	"pixeltactics.com/match/src/matches"
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
	playerSession, isPlayerInSession := repo.playerSession.Load(playerId)
	if isPlayerInSession {
		sessionOpponent, _ := playerSession.GetOpponentPlayerSync(playerId)
		if sessionOpponent.Id != opponentId {
			if playerSession.GetRunning() {
				return nil, errors.New("player is already on a session with another player")
			} else {
				repo.DeleteSession(playerSession.GetId())
			}
		} else {
			return nil, errors.New("session is already created")
		}
	}

	opponentSession, isOpponentInSession := repo.playerSession.Load(opponentId)
	if isOpponentInSession {
		sessionOpponent, _ := opponentSession.GetOpponentPlayerSync(opponentId)
		if sessionOpponent.Id != playerId {
			return nil, errors.New("opponent is already on a session with another player")
		} else {
			repo.playerSession.Store(playerId, opponentSession)
			opponentSession.RunSession()
			return opponentSession, nil
		}
	}

	newSessionId := uuid.New().String()
	_, ok := repo.sessions.Load(newSessionId)
	if ok {
		return nil, errors.New("duplicate session id")
	}

	matchMap, err := matches.GenerateMap()
	if err != nil {
		return nil, err
	}

	availableHeroList, err := matches.GetAvailableHeroes()
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

	player1Id, player2Id := session.GetPlayers()
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

var sessionRepository *SessionRepository = nil

func GetSessionRepository() *SessionRepository {
	if sessionRepository == nil {
		sessionRepository = &SessionRepository{
			sessions:      data_structures.NewSyncMap[string, *matches.Session](),
			playerSession: data_structures.NewSyncMap[string, *matches.Session](),
		}
	}
	return sessionRepository
}
