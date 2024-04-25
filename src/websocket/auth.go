package ws

import (
	"fortraiders.com/match/src/types"
	"fortraiders.com/match/src/utils"
)

type AuthMessageBody struct {
	playerToken string
}

type AuthHandler struct {
	clientHub       *ClientHub
	messageReceiver chan *MessageWithClient
}

func (handler *AuthHandler) AuthenticateClient(message *types.Message, client *Client) {
	var body AuthMessageBody
	utils.MapToObject(message.Body, &body)

	// TODO: Change this to JWT or smth
	playerId := body.playerToken

	handler.clientHub.RegisterPlayer(playerId, client)
}

func NewAuthHandler(clientHub *ClientHub) *AuthHandler {
	return &AuthHandler{
		clientHub:       clientHub,
		messageReceiver: make(chan *MessageWithClient),
	}
}

func (handler *AuthHandler) Run() {
	for {
		messageWithClient, ok := <-handler.messageReceiver
		if ok && messageWithClient.Message.Action == types.ACTION_AUTH {
			handler.AuthenticateClient(messageWithClient.Message, messageWithClient.Client)
		}
	}
}
