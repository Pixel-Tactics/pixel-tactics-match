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
	Move(client *messages.WebSocketMessager)
	Attack(client *messages.WebSocketMessager)
}

type ActionGatewayImpl struct {
	ActionService services.ActionService
	BaseGateway
}

func (gateway *ActionGatewayImpl) Move(client *messages.WebSocketMessager) {
	if client.ClientId == nil {
		client.SendBack(Error(errors.New("not authenticated")))
		return
	}

	var body dto.MoveRequest
	err := convert_utils.MapToObject(client.Message.Body, &body)
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

	err = gateway.ActionService.Move(*client.ClientId, body.Hero, body.Directions)
	if err != nil {
		client.SendBack(Error(err))
		return
	}
	// When action is executed, it will be sent back to user via event
}

func (gateway *ActionGatewayImpl) Attack(client *messages.WebSocketMessager) {
	if client.ClientId == nil {
		client.SendBack(Error(errors.New("not authenticated")))
		return
	}

	var body dto.AttackRequest
	err := convert_utils.MapToObject(client.Message.Body, &body)
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

	err = gateway.ActionService.Attack(*client.ClientId, body.SrcHero, body.DstHero)
	if err != nil {
		client.SendBack(Error(err))
		return
	}
	// When action is executed, it will be sent back to user via event
}

// func (gateway *SessionGatewayImpl) EndTurn(client *messages.WebSocketMessager) {
// 	playerIdInterface, ok := req.Message.Body["playerId"]
// 	if !ok {
// 		res.SendToClient(responses.ErrorMessage(errors.New("no player id")))
// 		return
// 	}
// 	playerId, ok := playerIdInterface.(string)
// 	if !ok {
// 		res.SendToClient(responses.ErrorMessage(errors.New("player id must be a string")))
// 		return
// 	}

// 	err := handler.matchService.EndTurn(playerId)
// 	if err != nil {
// 		res.SendToClient(responses.ErrorMessage(err))
// 		return
// 	}
// }

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
