package matches

import (
	"errors"

	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"
	"pixeltactics.com/match/src/utils"
)

type SessionRepository struct {
	sessions      *xsync.MapOf[string, *Session]
	playerSession *xsync.MapOf[string, *Session]
}

var sessionRepository *SessionRepository = nil

func (repo SessionRepository) GetSessionById(id string) *Session {
	value, ok := repo.sessions.Load(id)
	if !ok {
		return nil
	}
	return value
}

func (repo SessionRepository) GetSessionByPlayer(player *Player) *Session {
	value, ok := repo.playerSession.Load(player.Id)
	if !ok {
		return nil
	}
	return value
}

func (repo SessionRepository) CreateSession(player *Player, opponent *Player) (*Session, error) {
	playerSession, isPlayerInSession := repo.playerSession.Load(player.Id)
	if isPlayerInSession {
		sessionOpponent, _ := playerSession.GetOpponentPlayer(player)
		if sessionOpponent.Id != opponent.Id {
			return nil, errors.New("player is already on a session with another player")
		} else {
			return nil, errors.New("session is already created")
		}
	}

	opponentSession, isOpponentInSession := repo.playerSession.Load(opponent.Id)
	if isOpponentInSession {
		sessionOpponent, _ := opponentSession.GetOpponentPlayer(opponent)
		if sessionOpponent.Id != player.Id {
			return nil, errors.New("opponent is already on a session with another player")
		} else {
			repo.playerSession.Store(player.Id, opponentSession)
			opponentSession.Running = true
			return opponentSession, nil
		}
	}

	newSessionId := uuid.New().String()
	_, ok := repo.sessions.Load(newSessionId)
	if ok {
		return nil, errors.New("duplicate session id")
	}

	key, err := utils.GenerateSecureKey(32)
	if err != nil {
		return nil, err
	}

	matchMap, err := GenerateMap()
	if err != nil {
		return nil, err
	}

	availableHeroList, err := GetAvailableHeroes()
	if err != nil {
		return nil, err
	}

	newSession := &Session{
		Id:                newSessionId,
		Secret:            key,
		Player1:           player,
		Player2:           opponent,
		Running:           false,
		MatchMap:          matchMap,
		AvailableHeroList: availableHeroList,
	}

	repo.sessions.Store(newSessionId, newSession)
	repo.playerSession.Store(player.Id, newSession)
	return newSession, nil
}

func GetSessionRepository() *SessionRepository {
	if sessionRepository == nil {
		sessionRepository = &SessionRepository{
			sessions:      xsync.NewMapOf[string, *Session](),
			playerSession: xsync.NewMapOf[string, *Session](),
		}
	}
	return sessionRepository
}
