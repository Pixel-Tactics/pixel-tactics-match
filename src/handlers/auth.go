package handlers

import (
	"pixeltactics.com/match/src/types"
	"pixeltactics.com/match/src/utils"
)

type AuthMessageBody struct {
	PlayerToken string `json:"playerToken"`
}

type AuthHandler struct {
	Interaction chan *types.Interaction
}

func (handler *AuthHandler) AuthenticateClient(req *types.Request, res *types.Response) {
	var body AuthMessageBody
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	// TODO: Change this to JWT or smth
	playerId := body.PlayerToken
	if len(playerId) == 0 {
		res.SendToClient(&types.Message{
			Action: types.ACTION_ERROR,
			Body: map[string]interface{}{
				"status":  "failed",
				"message": "invalid player token",
			},
		})
		return
	}

	res.RegisterPlayer(playerId)
	res.SendToClient(&types.Message{
		Action: types.ACTION_FEEDBACK,
		Body: map[string]interface{}{
			"status":  "success",
			"message": "user registered on session",
		},
	})
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		Interaction: make(chan *types.Interaction),
	}
}

func (handler *AuthHandler) Run() {
	for {
		interaction, ok := <-handler.Interaction
		req := interaction.Request
		res := interaction.Response
		if ok && req.Message.Action == types.ACTION_AUTH {
			handler.AuthenticateClient(req, res)
		}
	}
}
