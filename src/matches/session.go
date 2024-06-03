package matches

import (
	"errors"
	"log"
	"sync"
	"time"
)

// 0 = Background
// 1 = Land
// 2 = Spawn
// 3 = Obstacle
type MatchMap struct {
	Structure [][]int `json:"structure"`
}

func (m *MatchMap) IsPointOpen(pos Point) bool {
	curValue := m.Structure[pos.x][pos.y]
	return curValue == 1 || curValue == 2
}

type Hero struct {
	Health         int   `json:"health"`
	Pos            Point `json:"pos"`
	lastMoveTurn   int   `json:"lastMoveTurn"`
	lastAttackTurn int   `json:"lastAttackTurn"`
	HeroTemplate   `json:"template"`
}

func (h *Hero) canMove(currentTurn int) bool {
	if h.lastAttackTurn >= currentTurn {
		return false
	} else if h.lastMoveTurn >= currentTurn {
		return false
	} else {
		return true
	}
}

func (h *Hero) canAttack(currentTurn int) bool {
	return h.lastAttackTurn < currentTurn
}

type Player struct {
	Id       string  `json:"id"`
	HeroList []*Hero `json:"heroList"`
}

func (p *Player) IsHeroExists(hero *Hero) bool {
	for _, curHero := range p.HeroList {
		if hero == curHero {
			return true
		}
	}
	return false
}

type Session struct {
	id                string
	player1           *Player
	player2           *Player
	running           bool
	currentTurn       int
	state             ISessionState
	matchMap          *MatchMap
	availableHeroList []string
	actionLog         []IAction
	lock              sync.Mutex
}

/* No Mutex Methods */
func (session *Session) getOpponentPlayer(playerId string) (*Player, error) {
	if playerId == session.player1.Id {
		return session.player2, nil
	} else if playerId == session.player2.Id {
		return session.player1, nil
	} else {
		return nil, errors.New("player in session")
	}
}

func (session *Session) getPlayerFromId(playerId string) *Player {
	if session.player1.Id == playerId {
		return session.player1
	} else if session.player2.Id == playerId {
		return session.player2
	} else {
		return nil
	}
}

func (session *Session) getHeroOnPlayer(playerId string, heroName string) (*Hero, error) {
	player := session.getPlayerFromId(playerId)
	if player == nil {
		return nil, errors.New("invalid player id")
	}
	for _, hero := range player.HeroList {
		if hero.GetName() == heroName {
			return hero, nil
		}
	}
	return nil, errors.New("invalid hero")
}

func (session *Session) isPointOpen(pos Point) bool {
	for _, hero := range session.player1.HeroList {
		if hero.Pos == pos {
			return false
		}
	}
	for _, hero := range session.player2.HeroList {
		if hero.Pos == pos {
			return false
		}
	}
	return session.matchMap.IsPointOpen(pos)
}

func (session *Session) checkWinner() string {
	cnt := 0
	for _, hero := range session.player1.HeroList {
		if hero.Health == 0 {
			cnt++
		}
	}
	if cnt == len(session.player1.HeroList) {
		return session.player1.Id
	}
	cnt = 0
	for _, hero := range session.player2.HeroList {
		if hero.Health == 0 {
			cnt++
		}
	}
	if cnt == len(session.player2.HeroList) {
		return session.player2.Id
	}
	return ""
}

func (session *Session) processEndResult() {
	log.Println("match " + session.id + " completed")
}

/* Mutex Methods */
func (session *Session) GetPlayers() (string, string) {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.player1.Id, session.player2.Id
}

func (session *Session) GetHeroOnPlayerSync(playerId string, heroName string) (Hero, error) {
	session.lock.Lock()
	defer session.lock.Unlock()
	heroPtr, err := session.getHeroOnPlayer(playerId, heroName)
	return *heroPtr, err
}

func (session *Session) GetOpponentPlayerSync(playerId string) (Player, error) {
	session.lock.Lock()
	defer session.lock.Unlock()
	playerPtr, err := session.getOpponentPlayer(playerId)
	return *playerPtr, err
}

func (session *Session) GetOpponentPlayerIdSync(playerId string) (string, error) {
	session.lock.Lock()
	defer session.lock.Unlock()
	opponent, err := session.getOpponentPlayer(playerId)
	if err != nil {
		return "", err
	}
	return opponent.Id, nil
}

func (session *Session) GetPlayerFromIdSync(playerId string) Player {
	session.lock.Lock()
	defer session.lock.Unlock()
	return *session.getPlayerFromId(playerId)
}

func (session *Session) RunSession() {
	session.lock.Lock()
	defer session.lock.Unlock()
	session.running = true
	session.state = &PreparationState{
		session:  session,
		deadline: time.Now().Add(preparationTime),
	}
}

func (session *Session) GetRunning() bool {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.running
}

func (session *Session) GetMatchMap() MatchMap {
	session.lock.Lock()
	defer session.lock.Unlock()
	return *session.matchMap
}

func (session *Session) GetAvailableHeroes() []string {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.availableHeroList
}

func (session *Session) PreparePlayer(playerId string, chosenHeroList []*Hero) error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.preparePlayer(playerId, chosenHeroList)
}

func (session *Session) StartBattle() error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.startBattle()
}

func (session *Session) ExecuteAction(action IAction) error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.executeAction(action)
}

func (session *Session) Forfeit(playerId string) error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.forfeit(playerId)
}

func (session *Session) GetData() map[string]interface{} {
	session.lock.Lock()
	defer session.lock.Unlock()
	return map[string]interface{}{
		"id":                session.id,
		"player1":           session.player1,
		"player2":           session.player2,
		"state":             session.state.getData(),
		"availableHeroList": session.availableHeroList,
		"matchMap":          session.matchMap,
		// "actionLog":         session.actionLog,
	}
}

func NewSession(sessionId string, player1Id string, player2Id string, matchMap *MatchMap, availableHeroList []string) *Session {
	player1 := &Player{
		Id:       player1Id,
		HeroList: []*Hero{},
	}
	player2 := &Player{
		Id:       player2Id,
		HeroList: []*Hero{},
	}
	newMatchMap := matchMap
	newAvailableHeroList := availableHeroList
	return &Session{
		id:                sessionId,
		player1:           player1,
		player2:           player2,
		running:           false,
		matchMap:          newMatchMap,
		availableHeroList: newAvailableHeroList,
	}
}

func GenerateMap() (*MatchMap, error) {
	structure := [][]int{
		{1, 1, 1, 1, 1},
		{1, 2, 2, 1, 1},
		{1, 1, 1, 1, 1},
		{1, 1, 2, 2, 1},
		{1, 1, 1, 1, 1},
	}
	newMap := MatchMap{
		Structure: structure,
	}
	return &newMap, nil
}

func GetAvailableHeroes() ([]string, error) {
	return []string{"knight"}, nil
}
