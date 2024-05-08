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
		utils.ErrorMessage(errors.New("invalid message body"))
		return
	}
	session, err := handler.MatchService.GetSession(body)
	if err != nil {
		utils.ErrorMessage(err)
		return
	}

	resBody, err := utils.ObjectToMap(session)
	if err != nil {
		utils.ErrorMessage(err)
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
		utils.ErrorMessage(err)
		return
	}

	session, err := handler.MatchService.CreateSession(body)
	if err != nil {
		utils.ErrorMessage(err)
		return
	}

	resBody, err := utils.ObjectToMap(session)
	if err != nil {
		utils.ErrorMessage(err)
		return
	}

	res.SendToClient(&types.Message{
		Action: types.ACTION_FEEDBACK,
		Body:   resBody,
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
