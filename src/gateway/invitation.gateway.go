package gateway

import (
	"log"

	"github.com/go-playground/validator/v10"
	"pixeltactics.com/match/src/dto"
	"pixeltactics.com/match/src/messages"
	"pixeltactics.com/match/src/services"
	convert_utils "pixeltactics.com/match/src/utils/convert"
)

type InvitationGateway interface {
	Invite(client *messages.WebSocketMessager)
}

type InvitationGatewayImpl struct {
	InvitationService services.InvitationService
	BaseGateway
}

func (gateway *InvitationGatewayImpl) Invite(client *messages.WebSocketMessager) {
	if client.ClientId == nil {
		client.SendBack(Error(ErrNotAuthenticated))
		return
	}

	var body dto.InvitationRequest
	err := convert_utils.MapToObject(client.Message.Body, &body)
	if err != nil {
		client.SendBack(Error(err))
		return
	}

	err = gateway.Validator.Struct(body)
	if err != nil {
		log.Println(err)
		client.SendBack(Error(ErrInvalidInput))
		return
	}

	isMutual, err := gateway.InvitationService.Invite(*client.ClientId, body.PlayerId)
	if err != nil {
		log.Println(err)
		client.SendBack(Error(err))
		return
	}
	if isMutual {
		// Session event will be sent instead
		return
	}

	client.SendBack(&messages.Message{
		Type: client.Message.Type,
		Body: map[string]interface{}{
			"message": "successfully sent an invitation...",
			"success": true,
		},
	})
}

func NewInvitationGateway(
	invitationService services.InvitationService,
	validator *validator.Validate,
) InvitationGateway {
	return &InvitationGatewayImpl{
		InvitationService: invitationService,
		BaseGateway: BaseGateway{
			Validator: validator,
		},
	}
}
