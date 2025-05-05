package gateway

import (
	"encoding/json"
	"log"

	"github.com/go-playground/validator/v10"
	"pixeltactics.com/match/src/dto"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/messages"
	"pixeltactics.com/match/src/repositories"
	"pixeltactics.com/match/src/services"
)

const INVITE_REQUEST_EVENT = "INVITE_REQUEST_EVENT"

type InvitationGateway interface {
	Invite(data interface{}) error
	HasMessager
}

type InvitationGatewayImpl struct {
	InvitationService services.InvitationService
	EventManager      events.EventManager
	BaseGateway
}

func (gateway *InvitationGatewayImpl) Invite(data interface{}) error {
	dataBytes, ok := data.([]byte)
	if !ok {
		log.Println("invalid event class: ")
		panic(data)
	}

	var body dto.InvitationRequest
	err := json.Unmarshal(dataBytes, &body)
	if err != nil {
		return err
	}

	err = gateway.Validator.Struct(body)
	if err != nil {
		return err
	}

	_, err = gateway.InvitationService.Invite(body.Id, body.PlayerId, body.OpponentId)
	if err != nil && err == repositories.ErrDuplicated {
		return nil
	} else if err != nil {
		return err
	}
	return nil
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
	gateway.EventManager.On(INVITE_REQUEST_EVENT, gateway.Invite)
	return gateway
}
