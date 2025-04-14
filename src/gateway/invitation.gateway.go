package gateway

import (
	"log"

	"github.com/go-playground/validator/v10"
	"pixeltactics.com/match/src/dto"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/messages"
	"pixeltactics.com/match/src/services"
	convert_utils "pixeltactics.com/match/src/utils/convert"
)

type InvitationGateway interface {
	Invite(client *messages.WebSocketMessager)
	HasMessager
}

type InvitationGatewayImpl struct {
	InvitationService services.InvitationService
	EventManager      events.EventManager
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

func (gateway *InvitationGatewayImpl) NotifyInvite(data interface{}) error {
	event, ok := data.(*services.InviteEvent)
	if !ok {
		log.Println("invalid event class: ")
		panic(data)
	}
	gateway.Messager.Send(event.DstPlayerId, &messages.Message{
		Type: TYPE_INVITE_SESSION,
		Body: map[string]interface{}{
			"playerId": event.SrcPlayerId,
		},
	})
	return nil
}

func NewInvitationGateway(
	invitationService services.InvitationService,
	validator *validator.Validate,
	eventManager events.EventManager,
) InvitationGateway {
	gateway := &InvitationGatewayImpl{
		InvitationService: invitationService,
		EventManager:      eventManager,
		BaseGateway: BaseGateway{
			Validator: validator,
		},
	}
	gateway.EventManager.On(services.INVITE_EVENT, gateway.NotifyInvite)
	return gateway
}
