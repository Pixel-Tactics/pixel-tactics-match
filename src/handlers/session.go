package handlers

import (
	"errors"

	"pixeltactics.com/match/src/notifiers"
	"pixeltactics.com/match/src/services"
	"pixeltactics.com/match/src/types"
	"pixeltactics.com/match/src/utils"
)

type SessionHandler struct {
	matchService services.MatchService
	Interaction  chan *types.Interaction
}

func (handler *SessionHandler) GetSession(req *types.Request, res *types.Response) {
	var body services.GetSessionRequestDTO
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}
	session, err := handler.matchService.GetSession(body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	res.SendToClient(&types.Message{
		Action:     types.ACTION_FEEDBACK,
		Identifier: req.Message.Identifier,
		Body:       session,
	})
}

func (handler *SessionHandler) CreateSession(req *types.Request, res *types.Response) {
	var body services.CreateSessionRequestDTO
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	if body.PlayerId == body.OpponentId {
		res.SendToClient(utils.ErrorMessage(errors.New("invalid opponent")))
		return
	}

	session, err := handler.matchService.CreateSession(body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	res.SendToClient(&types.Message{
		Action: types.ACTION_FEEDBACK,
		Body: map[string]interface{}{
			"success": true,
		},
	})

	if !session.GetRunning() {
		res.NotifyOtherClient(body.OpponentId, &types.Message{
			Action: types.ACTION_INVITE_SESSION,
			Body: map[string]interface{}{
				"playerId": body.PlayerId,
			},
		})
	} else {
		sessionMap := session.GetData()
		res.NotifyOtherClient(body.OpponentId, &types.Message{
			Action: types.ACTION_START_SESSION,
			Body: map[string]interface{}{
				"opponentId": body.PlayerId,
				"session":    sessionMap,
			},
		})
		res.NotifyClient(&types.Message{
			Action: types.ACTION_START_SESSION,
			Body: map[string]interface{}{
				"opponentId": body.OpponentId,
				"session":    sessionMap,
			},
		})
	}
}

func (handler *SessionHandler) PreparePlayer(req *types.Request, res *types.Response) {
	var body services.PreparePlayerRequestDTO
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	_, err = handler.matchService.PreparePlayer(body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	res.SendToClient(&types.Message{
		Action: types.ACTION_FEEDBACK,
		Body: map[string]interface{}{
			"success": true,
		},
	})
}

func (handler *SessionHandler) ExecuteAction(req *types.Request, res *types.Response) {
	var body services.ExecuteActionRequestDTO
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	// opponentId, err := handler.matchService.GetOpponentId(body.PlayerId)
	// if err != nil {
	// 	res.SendToClient(utils.ErrorMessage(err))
	// 	return
	// }

	err = handler.matchService.ExecuteAction(body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	// res.NotifyClient(&types.Message{
	// 	Action: types.ACTION_APPLY_ACTION,
	// 	Body: map[string]interface{}{
	// 		"actionName":     body.ActionName,
	// 		"actionSpecific": body.ActionSpecific,
	// 	},
	// })

	// res.NotifyOtherClient(opponentId, &types.Message{
	// 	Action: types.ACTION_APPLY_ACTION,
	// 	Body: map[string]interface{}{
	// 		"actionName":     body.ActionName,
	// 		"actionSpecific": body.ActionSpecific,
	// 	},
	// })
}

func (handler *SessionHandler) EndTurn(req *types.Request, res *types.Response) {
	playerIdInterface, ok := req.Message.Body["playerId"]
	if !ok {
		res.SendToClient(utils.ErrorMessage(errors.New("no player id")))
		return
	}
	playerId, ok := playerIdInterface.(string)
	if !ok {
		res.SendToClient(utils.ErrorMessage(errors.New("player id must be a string")))
		return
	}

	err := handler.matchService.EndTurn(playerId)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}
}

func (handler *SessionHandler) GetServerTime(req *types.Request, res *types.Response) {
	curTime := float64(handler.matchService.GetServerTime().UnixMilli())
	resTime := curTime / 1000.0
	res.SendToClient(&types.Message{
		Action: types.ACTION_FEEDBACK,
		Body: map[string]interface{}{
			"localTime":  req.Message.Body["localTime"],
			"serverTime": resTime,
		},
	})
}

func (handler *SessionHandler) Run() {
	for {
		interaction, ok := <-handler.Interaction
		req := interaction.Request
		res := interaction.Response
		if ok {
			if req.Message.Action == types.ACTION_GET_SESSION {
				handler.GetSession(req, res)
			} else if req.Message.Action == types.ACTION_CREATE_SESSION {
				handler.CreateSession(req, res)
			} else if req.Message.Action == types.ACTION_SERVER_TIME {
				handler.GetServerTime(req, res)
			} else if req.Message.Action == types.ACTION_PREPARE_PLAYER {
				handler.PreparePlayer(req, res)
			} else if req.Message.Action == types.ACTION_EXECUTE_ACTION {
				handler.ExecuteAction(req, res)
			} else if req.Message.Action == types.ACTION_END_TURN {
				handler.EndTurn(req, res)
			}
			handler.handleNotifierChannel(res)
		}
	}
}

func (handler *SessionHandler) handleNotifierChannel(res *types.Response) {
	notifier := notifiers.GetSessionNotifier()
	for {
		isBreak := false
		select {
		case msg, ok := <-notifier.SendChannel:
			if ok {
				playerId := msg.PlayerId
				message := msg.Message
				res.NotifyOtherClient(playerId, &message)
			}
		default:
			isBreak = true
		}

		if isBreak {
			break
		}
	}
}

func NewSessionHandler() *SessionHandler {
	return &SessionHandler{
		matchService: *services.NewMatchService(),
		Interaction:  make(chan *types.Interaction, 256),
	}
}
