package handlers

import (
	"errors"

	"pixeltactics.com/match/src/notifiers"
	"pixeltactics.com/match/src/services"
	"pixeltactics.com/match/src/utils"
	ws_types "pixeltactics.com/match/src/websocket/types"
)

type SessionHandler struct {
	matchService services.MatchService
	Interaction  chan *ws_types.Interaction
}

func (handler *SessionHandler) GetIsPlayerInSession(req *ws_types.Request, res *ws_types.Response) {
	var body services.PlayerIdDTO
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	_, err = handler.matchService.GetPlayerSession(body.PlayerId)
	res.SendToClient(&ws_types.Message{
		Action:     ws_types.ACTION_FEEDBACK,
		Identifier: req.Message.Identifier,
		Body: map[string]interface{}{
			"inSession": err == nil,
		},
	})
}

func (handler *SessionHandler) GetSession(req *ws_types.Request, res *ws_types.Response) {
	var body services.PlayerIdDTO
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	session, err := handler.matchService.GetPlayerSession(body.PlayerId)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	res.SendToClient(&ws_types.Message{
		Action:     ws_types.ACTION_FEEDBACK,
		Identifier: req.Message.Identifier,
		Body:       session,
	})
}

func (handler *SessionHandler) CreateSession(req *ws_types.Request, res *ws_types.Response) {
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

	res.SendToClient(&ws_types.Message{
		Action: ws_types.ACTION_FEEDBACK,
		Body: map[string]interface{}{
			"success": true,
		},
	})

	if !session.GetRunningSync() {
		res.NotifyOtherClient(body.OpponentId, &ws_types.Message{
			Action: ws_types.ACTION_INVITE_SESSION,
			Body: map[string]interface{}{
				"playerId": body.PlayerId,
			},
		})
	} else {
		sessionMap := session.GetDataSync()
		res.NotifyOtherClient(body.OpponentId, &ws_types.Message{
			Action: ws_types.ACTION_START_SESSION,
			Body: map[string]interface{}{
				"opponentId": body.PlayerId,
				"session":    sessionMap,
			},
		})
		res.NotifyClient(&ws_types.Message{
			Action: ws_types.ACTION_START_SESSION,
			Body: map[string]interface{}{
				"opponentId": body.OpponentId,
				"session":    sessionMap,
			},
		})
	}
}

func (handler *SessionHandler) PreparePlayer(req *ws_types.Request, res *ws_types.Response) {
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

	res.SendToClient(&ws_types.Message{
		Action: ws_types.ACTION_FEEDBACK,
		Body: map[string]interface{}{
			"success": true,
		},
	})
}

func (handler *SessionHandler) ExecuteAction(req *ws_types.Request, res *ws_types.Response) {
	var body services.ExecuteActionRequestDTO
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	opponentId, err := handler.matchService.GetOpponentId(body.PlayerId)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	action, err := handler.matchService.ExecuteAction(body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	res.NotifyClient(&ws_types.Message{
		Action: ws_types.ACTION_APPLY_ACTION,
		Body: map[string]interface{}{
			"actionName":     body.ActionName,
			"actionSpecific": action,
		},
	})
	res.NotifyOtherClient(opponentId, &ws_types.Message{
		Action: ws_types.ACTION_APPLY_ACTION,
		Body: map[string]interface{}{
			"actionName":     body.ActionName,
			"actionSpecific": action,
		},
	})
}

func (handler *SessionHandler) EndTurn(req *ws_types.Request, res *ws_types.Response) {
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

func (handler *SessionHandler) GetServerTime(req *ws_types.Request, res *ws_types.Response) {
	curTime := float64(handler.matchService.GetServerTime().UnixMilli())
	resTime := curTime / 1000.0
	res.SendToClient(&ws_types.Message{
		Action: ws_types.ACTION_FEEDBACK,
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
			if req.Message.Action == ws_types.ACTION_IS_IN_SESSION {
				handler.GetIsPlayerInSession(req, res)
			} else if req.Message.Action == ws_types.ACTION_GET_SESSION {
				handler.GetSession(req, res)
			} else if req.Message.Action == ws_types.ACTION_CREATE_SESSION {
				handler.CreateSession(req, res)
			} else if req.Message.Action == ws_types.ACTION_SERVER_TIME {
				handler.GetServerTime(req, res)
			} else if req.Message.Action == ws_types.ACTION_PREPARE_PLAYER {
				handler.PreparePlayer(req, res)
			} else if req.Message.Action == ws_types.ACTION_EXECUTE_ACTION {
				handler.ExecuteAction(req, res)
			} else if req.Message.Action == ws_types.ACTION_END_TURN {
				handler.EndTurn(req, res)
			}
			handler.handleNotifierChannel(res)
		}
	}
}

func (handler *SessionHandler) handleNotifierChannel(res *ws_types.Response) {
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
		Interaction:  make(chan *ws_types.Interaction, 256),
	}
}
