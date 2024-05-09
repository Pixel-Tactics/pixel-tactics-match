package handlers

import (
	"errors"

	"pixeltactics.com/match/src/matches"
	"pixeltactics.com/match/src/types"
	"pixeltactics.com/match/src/utils"
)

type SessionHandler struct {
	MatchService matches.MatchService
	Interaction  chan *types.Interaction
}

func (handler *SessionHandler) GetSession(req *types.Request, res *types.Response) {
	var body matches.GetSessionRequestDTO
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(req.Message.Identifier, errors.New("invalid message body")))
		return
	}
	session, err := handler.MatchService.GetSession(body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(req.Message.Identifier, err))
		return
	}

	resBody, err := utils.ObjectToMap(session)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(req.Message.Identifier, err))
		return
	}

	res.SendToClient(&types.Message{
		Action: types.ACTION_FEEDBACK,
		Body:   resBody,
	})
}

func (handler *SessionHandler) CreateSession(req *types.Request, res *types.Response) {
	var body matches.CreateSessionRequestDTO
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(req.Message.Identifier, errors.New("invalid message body")))
		return
	}

	if body.PlayerId == body.OpponentId {
		res.SendToClient(utils.ErrorMessage(req.Message.Identifier, errors.New("invalid opponent")))
		return
	}

	session, err := handler.MatchService.CreateSession(body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(req.Message.Identifier, err))
		return
	}

	sessionMap, err := utils.ObjectToMap(session)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(req.Message.Identifier, err))
		return
	}

	if !session.Running {
		res.SendToOtherClient(body.OpponentId, &types.Message{
			Action: types.ACTION_INVITE_SESSION,
			Body: map[string]interface{}{
				"playerId": body.PlayerId,
			},
		})
		res.SendToClient(&types.Message{
			Action: types.ACTION_FEEDBACK,
			Body: map[string]interface{}{
				"success": true,
			},
		})
	} else {
		res.SendToOtherClient(body.OpponentId, &types.Message{
			Action: types.ACTION_START_SESSION,
			Body: map[string]interface{}{
				"playerId": body.PlayerId,
				"session":  sessionMap,
			},
		})
		res.SendToClient(&types.Message{
			Action: types.ACTION_START_SESSION,
			Body:   sessionMap,
		})
	}
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
			}
		}
	}
}

func NewSessionHandler() *SessionHandler {
	return &SessionHandler{
		MatchService: matches.MatchService{},
		Interaction:  make(chan *types.Interaction),
	}
}
