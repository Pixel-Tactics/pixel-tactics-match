package gateway

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"pixeltactics.com/match/src/dto"
	"pixeltactics.com/match/src/messages"
	"pixeltactics.com/match/src/services"
	convert_utils "pixeltactics.com/match/src/utils/convert"
)

type AuthGateway interface {
	AuthenticateClient(client *messages.WebSocketMessager)
}

type AuthGatewayImpl struct {
	AuthService services.AuthService
	BaseGateway
}

func (gateway *AuthGatewayImpl) AuthenticateClient(client *messages.WebSocketMessager) {
	var body dto.Auth
	err := convert_utils.MapToObject(client.Message.Body, &body)
	if err != nil {
		client.SendBack(Error(err))
		return
	}

	err = gateway.Validator.Struct(body)
	if err != nil {
		client.SendBack(Error(errors.New("invalid input")))
		return
	}

	username, err := gateway.AuthService.GetUsernameFromToken(body.PlayerToken)
	if err != nil {
		client.SendBack(Error(err))
	} else {
		client.SetClientId(username)
		client.SendBack(&messages.Message{
			Type:       client.Message.Type,
			Identifier: client.Message.Identifier,
			Body: map[string]interface{}{
				"status":  "success",
				"message": "user registered on session as " + username,
			},
		})
	}
}

func NewAuthGateway(
	authService services.AuthService,
	validator *validator.Validate,
) AuthGateway {
	return &AuthGatewayImpl{
		AuthService: authService,
		BaseGateway: BaseGateway{
			Validator: validator,
		},
	}
}
