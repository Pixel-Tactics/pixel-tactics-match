package gateway

import (
	"errors"
	"log"

	"github.com/go-playground/validator/v10"
	"pixeltactics.com/match/src/dto"
	"pixeltactics.com/match/src/messages"
	"pixeltactics.com/match/src/services"
	convert_utils "pixeltactics.com/match/src/utils/convert"
)

type ActionGateway interface {
	Move(client messages.ClientMessager, message *messages.Message)
	Attack(client messages.ClientMessager, message *messages.Message)
	EndTurn(client messages.ClientMessager, message *messages.Message)
}

type ActionGatewayImpl struct {
	ActionService services.ActionService
	BaseGateway
}

func (gateway *ActionGatewayImpl) Move(client messages.ClientMessager, message *messages.Message) {
	var body dto.MoveRequest
	err := convert_utils.MapToObject(message.Data, &body)
	if err != nil {
		client.SendBack(Error(err))
		return
	}

	err = gateway.Validator.Struct(body)
	if err != nil {
		log.Println(err)
		client.SendBack(Error(errors.New("invalid input")))
		return
	}

	err = gateway.ActionService.Move(client.GetUsername(), body.Hero, body.Directions)
	if err != nil {
		client.SendBack(Error(err))
		return
	}
	// When action is executed, it will be sent back to user via event
}

func (gateway *ActionGatewayImpl) Attack(client messages.ClientMessager, message *messages.Message) {
	var body dto.AttackRequest
	err := convert_utils.MapToObject(message.Data, &body)
	if err != nil {
		client.SendBack(Error(err))
		return
	}

	err = gateway.Validator.Struct(body)
	if err != nil {
		log.Println(err)
		client.SendBack(Error(errors.New("invalid input")))
		return
	}

	err = gateway.ActionService.Attack(client.GetUsername(), body.SrcHero, body.DstHero)
	if err != nil {
		client.SendBack(Error(err))
		return
	}
	// When action is executed, it will be sent back to user via event
}

func (gateway *ActionGatewayImpl) EndTurn(client messages.ClientMessager, message *messages.Message) {
	err := gateway.ActionService.EndTurn(client.GetUsername())
	if err != nil {
		client.SendBack(Error(err))
		return
	}
	// When action is executed, it will be sent back to user via event
}

func NewActionGateway(
	actionService services.ActionService,
	validator *validator.Validate,
) ActionGateway {
	return &ActionGatewayImpl{
		ActionService: actionService,
		BaseGateway: BaseGateway{
			Validator: validator,
		},
	}
}
