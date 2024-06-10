package matches

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"pixeltactics.com/match/src/messaging"
	"pixeltactics.com/match/src/notifiers"
	"pixeltactics.com/match/src/utils"
)

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
	lock              *sync.Mutex
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
		if hero.Pos.Equals(pos) {
			return false
		}
	}
	for _, hero := range session.player2.HeroList {
		if hero.Pos.Equals(pos) {
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

func (session *Session) getLastAction() (IAction, bool) {
	if len(session.actionLog) == 0 {
		return nil, false
	}
	return session.actionLog[len(session.actionLog)-1], true
}

func (session *Session) processEndResult() {
	winner := session.checkWinner()
	if winner == "" {
		return
	}

	body := map[string]interface{}{
		"username1":  session.player1.Id,
		"username2":  session.player2.Id,
		"isUser1Win": (winner == session.player1.Id),
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		panic("invalid format on process end result json")
	}

	publisher := messaging.GetPublisher()
	publisher.Publish(&messaging.PublisherMessage{
		Exchange:   "matches",
		RoutingKey: "users",
		Body:       string(encoded),
	})

	log.Println("match " + session.id + " completed")
}

func (session *Session) createActionLog(actionName string, actionBody map[string]interface{}) (IAction, error) {
	if actionName == "move" {
		var action MoveLogData
		err := utils.MapToObject(actionBody, &action)
		if err != nil {
			return nil, err
		}

		hero, err := session.getHeroOnPlayer(action.PlayerId, action.HeroName)
		if err != nil {
			return nil, err
		}

		return &MoveLog{
			srcHero:       hero,
			directionList: action.DirectionList,
		}, nil
	} else if actionName == "attack" {
		var action AttackLogData
		err := utils.MapToObject(actionBody, &action)
		if err != nil {
			return nil, err
		}

		hero, err := session.getHeroOnPlayer(action.PlayerId, action.HeroName)
		if err != nil {
			return nil, err
		}

		opponent, err := session.getOpponentPlayer(action.PlayerId)
		if err != nil {
			return nil, err
		}

		target, err := session.getHeroOnPlayer(opponent.Id, action.TargetName)
		if err != nil {
			return nil, err
		}

		return &AttackLog{
			srcHero: hero,
			trgHero: target,
		}, nil
	}
	return nil, errors.New("invalid action name")
}

func (session *Session) getData() map[string]interface{} {
	actionLogData := []map[string]interface{}{}
	for _, actionLog := range session.actionLog {
		actionLogData = append(actionLogData, actionLog.getData())
	}
	return map[string]interface{}{
		"id":                session.id,
		"player1":           session.player1.getData(),
		"player2":           session.player2.getData(),
		"state":             session.state.getData(),
		"availableHeroList": session.availableHeroList,
		"matchMap":          session.matchMap,
		"actionLog":         actionLogData,
	}
}

func (session *Session) changeState(newState ISessionState) {
	_, ok1 := session.state.(*Player1TurnState)
	_, ok2 := session.state.(*Player2TurnState)
	if ok1 || ok2 {
		session.currentTurn += 1
	}

	session.state = newState
	notifier := notifiers.GetSessionNotifier()
	notifier.NotifyChangeState(session.player1.Id, session.getData())
	notifier.NotifyChangeState(session.player2.Id, session.getData())
	_, ok := session.state.(*EndState)
	if ok {
		session.processEndResult()
	}
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

func (session *Session) PreparePlayer(playerId string, chosenHeroList []HeroTemplate) error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.preparePlayer(playerId, chosenHeroList)
}

func (session *Session) StartBattle() error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.startBattle()
}

func (session *Session) ExecuteAction(actionName string, actionBody map[string]interface{}) error {
	session.lock.Lock()
	defer session.lock.Unlock()
	action, err := session.createActionLog(actionName, actionBody)
	if err != nil {
		return err
	}
	err = session.state.executeAction(action)
	if err != nil {
		return err
	}

	notifier := notifiers.GetSessionNotifier()
	notifier.NotifyAction(session.player1.Id, action.getName(), action.getData())
	notifier.NotifyAction(session.player2.Id, action.getName(), action.getData())
	return nil
}

func (session *Session) EndTurn(playerId string) error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.endTurn(playerId)
}

func (session *Session) Forfeit(playerId string) error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.forfeit(playerId)
}

func (session *Session) GetData() map[string]interface{} {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.getData()
}

func NewSession(sessionId string, player1Id string, player2Id string, matchMap *MatchMap, availableHeroList []string) *Session {
	newMatchMap := matchMap
	newAvailableHeroList := availableHeroList
	session := &Session{
		id:                sessionId,
		running:           false,
		matchMap:          newMatchMap,
		availableHeroList: newAvailableHeroList,
		lock:              new(sync.Mutex),
	}
	session.player1 = NewPlayer(player1Id, session)
	session.player2 = NewPlayer(player2Id, session)
	return session
}
