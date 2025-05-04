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
	Invite(client messages.ClientMessager, message *messages.Message)
	HasMessager
}

type InvitationGatewayImpl struct {
	InvitationService services.InvitationService
	EventManager      events.EventManager
	BaseGateway
}

func (gateway *InvitationGatewayImpl) Invite(client messages.ClientMessager, message *messages.Message) {
	var body dto.InvitationRequest
	err := convert_utils.MapToObject(message.Data, &body)
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

	isMutual, err := gateway.InvitationService.Invite(client.GetUsername(), body.PlayerId)
	if err != nil {
		log.Println(err)
		client.SendBack(Error(err))
		return
	}
	if isMutual {
		// Session event will be sent instead
		return
	}

	client.SendBack(messages.CreateMessage(message.Route, map[string]interface{}{
		"success": true,
	}, "successfully sent an invitation..."))
}

func (gateway *InvitationGatewayImpl) NotifyInvite(data interface{}) error {
	event, ok := data.(*services.InviteEvent)
	if !ok {
		log.Println("invalid event class: ")
		panic(data)
	}
	gateway.Messager.Send(event.DstPlayerId, messages.CreateMessage(
		TYPE_INVITE_SESSION,
		map[string]interface{}{
			"playerId": event.SrcPlayerId,
		},
		"",
	))
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
