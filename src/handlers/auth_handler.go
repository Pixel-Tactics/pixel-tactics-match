package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/types"
	"pixeltactics.com/match/src/utils"
)

type AuthMessageBody struct {
	PlayerToken string `json:"playerToken"`
}

type AuthHandler struct {
	Interaction      chan *types.Interaction
	successResponses chan *types.Interaction
	errorResponses   chan *types.Interaction
}

func (handler *AuthHandler) AuthenticateClient(req *types.Request, res *types.Response) {
	var body AuthMessageBody
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

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

	handler.sendAuthRequest(body.PlayerToken, res)
}

func (handler *AuthHandler) handleSuccess(req *types.Request, res *types.Response) {
	playerIdRaw, ok := req.Message.Body["playerId"]
	if !ok {
		log.Println(exceptions.InvalidJsonError())
		return
	}

	playerId, ok := playerIdRaw.(string)
	if !ok {
		log.Println(exceptions.InvalidJsonError())
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

func (handler *AuthHandler) handleError(req *types.Request, res *types.Response) {
	msgRaw, ok := req.Message.Body["message"]
	if !ok {
		log.Println(exceptions.InvalidJsonError())
		return
	}

	msg, ok := msgRaw.(string)
	if !ok {
		log.Println(exceptions.InvalidJsonError())
		return
	}

	res.SendToClient(utils.ErrorMessage(errors.New(msg)))
}

func (handler *AuthHandler) sendAuthRequest(playerToken string, res *types.Response) {
	go func() {
		sendError := func() {
			handler.errorResponses <- &types.Interaction{
				Request: &types.Request{
					Message: &types.Message{
						Body: map[string]interface{}{
							"message": "invalid token",
						},
					},
				},
				Response: res,
			}
		}

		host := os.Getenv("USER_MICROSERVICE_URL")
		client := &http.Client{}
		req, err := http.NewRequest("GET", host+"/auth/current", nil)
		if err != nil {
			log.Println("ERROR REQUEST MAKE")
			sendError()
		}
		req.Header.Set("Authorization", "Bearer "+playerToken)
		resp, err := client.Do(req)
		if err != nil {
			log.Println("ERROR REQUEST SEND")
			sendError()
		}

		log.Println("GOT RESPONSE")

		if resp.StatusCode == 200 {
			log.Println("RESPONSE 200")
			jsonBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				sendError()
			}

			var body map[string]string
			json.Unmarshal(jsonBytes, &body)

			playerId, ok := body["username"]
			if !ok {
				sendError()
			}

			handler.successResponses <- &types.Interaction{
				Request: &types.Request{
					Message: &types.Message{
						Body: map[string]interface{}{
							"playerId": playerId,
						},
					},
				},
				Response: res,
			}
		} else {
			log.Println("RESPONSE ERR")
			sendError()
		}
	}()
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		Interaction:      make(chan *types.Interaction, 256),
		successResponses: make(chan *types.Interaction, 256),
		errorResponses:   make(chan *types.Interaction, 256),
	}
}

func (handler *AuthHandler) Run() {
	for {
		select {
		case interaction := <-handler.Interaction:
			if interaction.Request.Message.Action == types.ACTION_AUTH {
				handler.AuthenticateClient(interaction.Request, interaction.Response)
			}
		case successResp := <-handler.successResponses:
			log.Println("TEST")
			handler.handleSuccess(successResp.Request, successResp.Response)
		case errorResp := <-handler.errorResponses:
			log.Println("TOWST")
			handler.handleError(errorResp.Request, errorResp.Response)
		}
	}
}
