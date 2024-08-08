package matches

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"pixeltactics.com/match/src/integrations/messaging"
	matches_actions "pixeltactics.com/match/src/matches/actions"
	matches_heroes "pixeltactics.com/match/src/matches/heroes"
	matches_interfaces "pixeltactics.com/match/src/matches/interfaces"
	matches_maps "pixeltactics.com/match/src/matches/maps"
	matches_physics "pixeltactics.com/match/src/matches/physics"
	matches_players "pixeltactics.com/match/src/matches/players"
	"pixeltactics.com/match/src/notifiers"
	"pixeltactics.com/match/src/utils"
)

type Session struct {
	id                string
	player1           *matches_players.Player
	player2           *matches_players.Player
	running           bool
	currentTurn       int
	state             ISessionState
	matchMap          *matches_maps.MatchMap
	availableHeroList []string
	actionLog         []matches_interfaces.IAction
	lock              *sync.Mutex
}

/* No Mutex Methods */
func (session *Session) getOpponentPlayer(playerId string) (*matches_players.Player, error) {
	if playerId == session.player1.Id {
		return session.player2, nil
	} else if playerId == session.player2.Id {
		return session.player1, nil
	} else {
		return nil, errors.New("player in session")
	}
}

func (session *Session) getPlayerFromId(playerId string) *matches_players.Player {
	if session.player1.Id == playerId {
		return session.player1
	} else if session.player2.Id == playerId {
		return session.player2
	} else {
		return nil
	}
}

func (session *Session) getHeroOnPlayer(playerId string, heroName string) (*matches_heroes.Hero, error) {
	player := session.getPlayerFromId(playerId)
	if player == nil {
		return nil, errors.New("invalid player id")
	}
	for _, hero := range player.HeroList {
		heroConc, ok := hero.(*matches_heroes.Hero)
		if ok && heroConc.GetName() == heroName {
			return heroConc, nil
		}
	}
	return nil, errors.New("invalid hero")
}

func (session *Session) IsPointOpen(pos matches_physics.Point) bool {
	for _, hero := range session.player1.HeroList {
		if hero.GetPos().Equals(pos) {
			return false
		}
	}
	for _, hero := range session.player2.HeroList {
		if hero.GetPos().Equals(pos) {
			return false
		}
	}
	return session.matchMap.IsPointOpen(pos)
}

func (session *Session) checkWinner() string {
	cntDead := 0
	for _, hero := range session.player1.HeroList {
		if hero.GetHealth() == 0 {
			cntDead++
		}
	}
	if cntDead == len(session.player1.HeroList) {
		return session.player2.Id
	}
	cntDead = 0
	for _, hero := range session.player2.HeroList {
		if hero.GetHealth() == 0 {
			cntDead++
		}
	}
	if cntDead == len(session.player2.HeroList) {
		return session.player1.Id
	}
	return ""
}

func (session *Session) GetLastAction() (matches_interfaces.IAction, bool) {
	if len(session.actionLog) == 0 {
		return nil, false
	}
	return session.actionLog[len(session.actionLog)-1], true
}

func (session *Session) processEndResult(endState *EndState) {
	session.running = false
	winner := endState.winnerId
	if winner == "draw" {
		return
	}

	body := map[string]interface{}{
		"matchId":    session.id,
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

func (session *Session) createActionLog(actionName string, actionBody map[string]interface{}) (matches_interfaces.IAction, error) {
	if actionName == "move" {
		var action matches_actions.MoveLogData
		err := utils.MapToObject(actionBody, &action)
		if err != nil {
			return nil, err
		}

		hero, err := session.getHeroOnPlayer(action.PlayerId, action.HeroName)
		if err != nil {
			return nil, err
		}

		return matches_actions.NewMoveLog(hero, action.DirectionList), nil
	} else if actionName == "attack" {
		var action matches_actions.AttackLogData
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

		return matches_actions.NewAttackLog(hero, target), nil
	}
	return nil, errors.New("invalid action name")
}

func (session *Session) getData() map[string]interface{} {
	actionLogData := []map[string]interface{}{}
	for i, actionLog := range session.actionLog {
		actionData := actionLog.GetData()
		actionData["order"] = i
		actionLogData = append(actionLogData, actionData)
	}
	return map[string]interface{}{
		"id":                session.id,
		"player1":           session.player1.GetData(),
		"player2":           session.player2.GetData(),
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
	timedState, isTimed := newState.(ITimedState)
	endState, ok := session.state.(*EndState)
	if ok {
		session.processEndResult(endState)
	} else if isTimed {
		time.AfterFunc(time.Until(timedState.getDeadline()), timedState.expire)
	}

	notifier := notifiers.GetSessionNotifier()
	notifier.NotifyChangeState(session.player1.Id, session.getData())
	notifier.NotifyChangeState(session.player2.Id, session.getData())
}

func (session *Session) GetMatchMap() *matches_maps.MatchMap {
	return session.matchMap
}

func (session *Session) GetCurrentTurn() int {
	return session.currentTurn
}

/* Mutex Methods */
func (session *Session) GetPlayersSync() (string, string) {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.player1.Id, session.player2.Id
}

func (session *Session) GetHeroOnPlayerSync(playerId string, heroName string) (matches_heroes.Hero, error) {
	session.lock.Lock()
	defer session.lock.Unlock()
	heroPtr, err := session.getHeroOnPlayer(playerId, heroName)
	return *heroPtr, err
}

func (session *Session) GetOpponentPlayerSync(playerId string) (matches_players.Player, error) {
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

func (session *Session) GetPlayerFromIdSync(playerId string) matches_players.Player {
	session.lock.Lock()
	defer session.lock.Unlock()
	return *session.getPlayerFromId(playerId)
}

func (session *Session) RunSessionSync() {
	session.lock.Lock()
	defer session.lock.Unlock()
	session.running = true
	newState := &PreparationState{
		session:  session,
		deadline: time.Now().Add(preparationTime),
	}
	session.state = newState
	time.AfterFunc(time.Until(newState.deadline), newState.expire)
}

func (session *Session) GetRunningSync() bool {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.running
}

func (session *Session) GetEndedSync() bool {
	session.lock.Lock()
	defer session.lock.Unlock()
	_, ok := session.state.(*EndState)
	return ok
}

func (session *Session) GetMatchMapSync() matches_maps.MatchMap {
	session.lock.Lock()
	defer session.lock.Unlock()
	return *session.GetMatchMap()
}

func (session *Session) GetAvailableHeroesSync() []string {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.availableHeroList
}

func (session *Session) PreparePlayerSync(playerId string, chosenHeroList []matches_interfaces.HeroTemplate) error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.preparePlayer(playerId, chosenHeroList)
}

func (session *Session) StartBattleSync() error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.startBattle()
}

func (session *Session) ExecuteActionSync(actionName string, actionBody map[string]interface{}) error {
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
	return nil
}

func (session *Session) EndTurnSync(playerId string) error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.endTurn(playerId)
}

func (session *Session) ForfeitSync(playerId string) error {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.state.forfeit(playerId)
}

func (session *Session) GetIdSync() string {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.id
}

func (session *Session) GetDataSync() map[string]interface{} {
	session.lock.Lock()
	defer session.lock.Unlock()
	return session.getData()
}

func NewSession(sessionId string, player1Id string, player2Id string, matchMap *matches_maps.MatchMap, availableHeroList []string) *Session {
	newMatchMap := matchMap
	newAvailableHeroList := availableHeroList
	session := &Session{
		id:                sessionId,
		running:           false,
		matchMap:          newMatchMap,
		availableHeroList: newAvailableHeroList,
		lock:              new(sync.Mutex),
	}
	session.player1 = matches_players.NewPlayer(player1Id, make([]matches_interfaces.IHero, 0), session)
	session.player2 = matches_players.NewPlayer(player2Id, make([]matches_interfaces.IHero, 0), session)
	return session
}
