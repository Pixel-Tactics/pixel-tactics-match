package gateway

import (
	"errors"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"pixeltactics.com/match/src/dto"
	"pixeltactics.com/match/src/events"
	"pixeltactics.com/match/src/messages"
	"pixeltactics.com/match/src/services"
	convert_utils "pixeltactics.com/match/src/utils/convert"
)

type SessionGateway interface {
	GetIsPlayerInSession(client messages.ClientMessager, message *messages.Message)
	GetSession(client messages.ClientMessager, message *messages.Message)
	PreparePlayer(client messages.ClientMessager, message *messages.Message)
	GetServerTime(client messages.ClientMessager, message *messages.Message)
	HasMessager
}

type SessionGatewayImpl struct {
	SessionService services.SessionService
	BaseGateway
}

func (gateway *SessionGatewayImpl) GetIsPlayerInSession(client messages.ClientMessager, message *messages.Message) {
	var body dto.PlayerId
	err := convert_utils.MapToObject(message.Data, &body)
	if err != nil {
		client.SendBack(Error(err))
		return
	}

	err = gateway.Validator.Struct(body)
	if err != nil {
		client.SendBack(Error(errors.New("invalid input")))
		return
	}

	session, err := gateway.SessionService.GetSessionByPlayerId(nil, body.PlayerId)
	if err != nil {
		client.SendBack(Error(err))
		return
	}

	client.SendBack(messages.CreateMessage(message.Route, map[string]interface{}{
		"inSession": session != nil,
	}, ""))
}

func (gateway *SessionGatewayImpl) GetSession(client messages.ClientMessager, message *messages.Message) {
	var body dto.PlayerId
	err := convert_utils.MapToObject(message.Data, &body)
	if err != nil {
		client.SendBack(Error(err))
		return
	}

	err = gateway.Validator.Struct(body)
	if err != nil {
		client.SendBack(Error(errors.New("invalid input")))
		return
	}

	session, err := gateway.SessionService.GetSessionByPlayerId(nil, body.PlayerId)
	if err != nil {
		client.SendBack(Error(err))
		return
	}
	if session == nil {
		client.SendBack(Error(errors.New("player is not in session")))
		return
	}

	response, _ := convert_utils.ObjectToMap(session)
	client.SendBack(messages.CreateMessage(message.Route, response, ""))
}

func (gateway *SessionGatewayImpl) PreparePlayer(client messages.ClientMessager, message *messages.Message) {
	var body dto.PreparePlayerRequest
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

	isStarted, err := gateway.SessionService.PreparePlayer(client.GetUsername(), body.ChosenHeroList)
	if err != nil {
		client.SendBack(Error(err))
		return
	}

	var msg string
	if isStarted {
		session, err := gateway.SessionService.GetSessionByPlayerId(nil, client.GetUsername())
		var notifyMessage string
		if err != nil {
			log.Println(err)
			notifyMessage = "Battle is started, but error on showing data"
		}

		response, err := gateway.SessionService.CompileSession(nil, session.Id)
		if err != nil {
			log.Println(err)
			notifyMessage = "Battle is started, but error on showing data"
		}

		otherId, _ := session.GetOtherPlayerId(client.GetUsername())
		client.Send(otherId, messages.CreateMessage(TYPE_START_BATTLE, map[string]interface{}{
			"opponentId": client.GetUsername(),
			"session":    response,
		}, notifyMessage))
		client.Send(client.GetUsername(), messages.CreateMessage(TYPE_START_BATTLE, map[string]interface{}{
			"opponentId": otherId,
			"session":    response,
		}, notifyMessage))

		msg = "Battle is started.."
		log.Println("Battle is started..")
	} else {
		msg = "Waiting for other player.."
	}

	client.SendBack(messages.CreateMessage(message.Route, map[string]interface{}{
		"success": true,
	}, msg))
}

func (gateway *SessionGatewayImpl) GetServerTime(client messages.ClientMessager, message *messages.Message) {
	var body dto.ServerTimeRequest
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

	curTime := float64(time.Now().UnixMilli())
	resTime := curTime / 1000.0

	client.SendBack(messages.CreateMessage(message.Route, map[string]interface{}{
		"localTime":  body.LocalTime,
		"serverTime": resTime,
	}, ""))
}

func (gateway *SessionGatewayImpl) NotifySessionCreated(data interface{}) error {
	event, ok := data.(*services.SessionCreatedEvent)
	if !ok {
		log.Println("invalid event class: ")
		panic(data)
	}
	for i, playerId := range event.PlayerIds {
		gateway.Messager.Send(playerId, messages.CreateMessage(
			TYPE_START_SESSION,
			map[string]interface{}{
				"opponentId": event.PlayerIds[(i+1)%2],
				"session":    event.Data,
			},
			"Session started...",
		))
	}
	return nil
}

func NewSessionGateway(
	sessionService services.SessionService,
	eventManager events.EventManager,
	validator *validator.Validate,
) SessionGateway {
	gateway := &SessionGatewayImpl{
		SessionService: sessionService,
		BaseGateway: BaseGateway{
			Validator: validator,
		},
	}
	eventManager.On(services.SESSION_CREATED_EVENT, gateway.NotifySessionCreated)
	return gateway
}
