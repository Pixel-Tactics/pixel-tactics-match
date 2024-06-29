package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"pixeltactics.com/match/src/exceptions"
	"pixeltactics.com/match/src/utils"
	ws_types "pixeltactics.com/match/src/websocket/types"
)

type AuthMessageBody struct {
	PlayerToken string `json:"playerToken"`
}

type AuthHandler struct {
	Interaction      chan *ws_types.Interaction
	successResponses chan *ws_types.Interaction
	errorResponses   chan *ws_types.Interaction
}

func (handler *AuthHandler) AuthenticateClient(req *ws_types.Request, res *ws_types.Response) {
	var body AuthMessageBody
	err := utils.MapToObject(req.Message.Body, &body)
	if err != nil {
		res.SendToClient(utils.ErrorMessage(err))
		return
	}

	playerId := body.PlayerToken
	if len(playerId) == 0 {
		res.SendToClient(&ws_types.Message{
			Action: ws_types.ACTION_ERROR,
			Body: map[string]interface{}{
				"status":  "failed",
				"message": "invalid player token",
			},
		})
		return
	}

	handler.sendAuthRequest(body.PlayerToken, res)
}

func (handler *AuthHandler) handleSuccess(req *ws_types.Request, res *ws_types.Response) {
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
	res.SendToClient(&ws_types.Message{
		Action: ws_types.ACTION_FEEDBACK,
		Body: map[string]interface{}{
			"status":  "success",
			"message": "user registered on session",
		},
	})
}

func (handler *AuthHandler) handleError(req *ws_types.Request, res *ws_types.Response) {
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

func (handler *AuthHandler) sendAuthRequest(playerToken string, res *ws_types.Response) {
	go func() {
		sendError := func() {
			handler.errorResponses <- &ws_types.Interaction{
				Request: &ws_types.Request{
					Message: &ws_types.Message{
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

		if resp.StatusCode == 200 {
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

			handler.successResponses <- &ws_types.Interaction{
				Request: &ws_types.Request{
					Message: &ws_types.Message{
						Body: map[string]interface{}{
							"playerId": playerId,
						},
					},
				},
				Response: res,
			}
		} else {
			sendError()
		}
	}()
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		Interaction:      make(chan *ws_types.Interaction, 256),
		successResponses: make(chan *ws_types.Interaction, 256),
		errorResponses:   make(chan *ws_types.Interaction, 256),
	}
}

func (handler *AuthHandler) Run() {
	for {
		select {
		case interaction := <-handler.Interaction:
			if interaction.Request.Message.Action == ws_types.ACTION_AUTH {
				handler.AuthenticateClient(interaction.Request, interaction.Response)
			}
		case successResp := <-handler.successResponses:
			handler.handleSuccess(successResp.Request, successResp.Response)
		case errorResp := <-handler.errorResponses:
			handler.handleError(errorResp.Request, errorResp.Response)
		}
	}
}
