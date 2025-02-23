package gateway

import (
	"errors"

	"pixeltactics.com/match/src/dto"
	"pixeltactics.com/match/src/services"
	convert_utils "pixeltactics.com/match/src/utils/convert"
)

type AuthGateway interface {
	AuthenticateClient(client *Client)
}

type AuthGatewayImpl struct {
	AuthService services.AuthService
	BaseGateway
}

func (gateway *AuthGatewayImpl) AuthenticateClient(client *Client) {
	var body dto.Auth
	err := convert_utils.MapToObject(client.Message.Body, &body)
	if err != nil {
		client.SendBack(Error(err))
		return
	}

	err = gateway.Validator.Struct(body)
	if err != nil {
		client.SendBack(Error(errors.New("invalid player token")))
		return
	}

	username, err := gateway.AuthService.GetUsernameFromToken(body.PlayerToken)
	if err != nil {
		client.SendBack(Error(errors.New("invalid player token")))
	} else {
		client.SetClientId(username)
		client.SendBack(&Message{
			Type:       client.Message.Type,
			Identifier: client.Message.Identifier,
			Body: map[string]interface{}{
				"status":  "success",
				"message": "user registered on session",
			},
		})
	}
}

// func NewAuthGateway() *AuthGateway {
// 	return &AuthGateway{
// 		Interaction:      make(chan *ws_types.Interaction, 256),
// 		successResponses: make(chan *ws_types.Interaction, 256),
// 		errorResponses:   make(chan *ws_types.Interaction, 256),
// 	}
// }

// func (gateway *AuthHandler) Run() {
// 	for {
// 		select {
// 		case interaction := <-handler.Interaction:
// 			if interaction.Request.Message.Action == ws_types.ACTION_AUTH {
// 				handler.AuthenticateClient(interaction.Request, interaction.Response)
// 			}
// 		case successResp := <-handler.successResponses:
// 			handler.handleSuccess(successResp.Request, successResp.Response)
// 		case errorResp := <-handler.errorResponses:
// 			handler.handleError(errorResp.Request, errorResp.Response)
// 		}
// 	}
// }
