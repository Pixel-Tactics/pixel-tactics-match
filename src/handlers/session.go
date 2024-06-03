package handlers

import (
	"errors"

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

	session, err := handler.matchService.GetPlayerSession(body.PlayerId)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	opponentId, err := handler.matchService.GetOpponentId(body.PlayerId)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	isStarted, err := handler.matchService.PreparePlayer(body)
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

	if isStarted {
		res.NotifyClient(&types.Message{
			Action: types.ACTION_START_BATTLE,
			Body: map[string]interface{}{
				"session": session,
			},
		})
		res.NotifyOtherClient(opponentId, &types.Message{
			Action: types.ACTION_START_BATTLE,
			Body: map[string]interface{}{
				"session": session,
			},
		})
	}
}

func (handler *SessionHandler) ExecuteAction(req *types.Request, res *types.Response) {
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

	err = handler.matchService.ExecuteAction(body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	res.SendToClient(&types.Message{
		Action:     types.ACTION_FEEDBACK,
		Identifier: req.Message.Identifier,
		Body: map[string]interface{}{
			"Success": true,
		},
	})

	res.NotifyOtherClient(opponentId, &types.Message{
		Action:     types.ACTION_ENEMY_ACTION,
		Identifier: req.Message.Identifier,
		Body: map[string]interface{}{
			"ActionName": body.ActionName,
			"Action":     body.ActionSpecific,
		},
	})
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
			}
		}
	}
}

func NewSessionHandler() *SessionHandler {
	return &SessionHandler{
		matchService: *services.NewMatchService(),
		Interaction:  make(chan *types.Interaction),
	}
}
