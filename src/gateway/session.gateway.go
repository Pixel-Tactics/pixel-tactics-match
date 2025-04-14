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
	GetIsPlayerInSession(client *messages.WebSocketMessager)
	GetSession(client *messages.WebSocketMessager)
	// CreateSession(client *messages.WebSocketMessager)
	PreparePlayer(client *messages.WebSocketMessager)
	GetServerTime(client *messages.WebSocketMessager)
	HasMessager
}

type SessionGatewayImpl struct {
	SessionService services.SessionService
	BaseGateway
}

func (gateway *SessionGatewayImpl) GetIsPlayerInSession(client *messages.WebSocketMessager) {
	var body dto.PlayerId
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

	session, err := gateway.SessionService.GetSessionByPlayerId(nil, body.PlayerId)
	if err != nil {
		client.SendBack(Error(err))
		return
	}
	client.SendBack(&messages.Message{
		Type:       client.Message.Type,
		Identifier: client.Message.Identifier,
		Body: map[string]interface{}{
			"inSession": session != nil,
		},
	})
}

func (gateway *SessionGatewayImpl) GetSession(client *messages.WebSocketMessager) {
	var body dto.PlayerId
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
	client.SendBack(&messages.Message{
		Type:       client.Message.Type,
		Identifier: client.Message.Identifier,
		Body:       response,
	})
}

// func (gateway *SessionGatewayImpl) CreateSession(client *messages.WebSocketMessager) {
// 	if client.ClientId == nil {
// 		client.SendBack(Error(errors.New("not authenticated")))
// 		return
// 	}

// 	var body dto.CreateSessionRequest
// 	err := convert_utils.MapToObject(client.Message.Body, &body)
// 	if err != nil {
// 		client.SendBack(Error(err))
// 		return
// 	}

// 	err = gateway.Validator.Struct(body)
// 	if err != nil {
// 		log.Println(err)
// 		client.SendBack(Error(errors.New("invalid input")))
// 		return
// 	}

// 	if *client.ClientId == body.OpponentId {
// 		client.SendBack(Error(errors.New("invalid opponent")))
// 		return
// 	}

// 	session, err := gateway.SessionService.CreateSession(*client.ClientId, body.OpponentId)
// 	if err != nil {
// 		client.SendBack(Error(err))
// 		return
// 	}

// 	client.SendBack(&messages.Message{
// 		Type: client.Message.Type,
// 		Body: map[string]interface{}{
// 			"success": true,
// 		},
// 	})

// 	if !session.IsRunning() {
// 		client.Send(body.OpponentId, &messages.Message{
// 			Type: TYPE_INVITE_SESSION,
// 			Body: map[string]interface{}{
// 				"playerId": *client.ClientId,
// 			},
// 		})
// 	} else {
// 		message := "successfully created session"
// 		response, err := gateway.SessionService.CompileSession(session.Id)
// 		if err != nil {
// 			log.Println(err)
// 			message = "successfully created session, but error on showing data"
// 		}
// 		client.Send(body.OpponentId, &messages.Message{
// 			Type: TYPE_START_SESSION,
// 			Body: map[string]interface{}{
// 				"message":    message,
// 				"opponentId": *client.ClientId,
// 				"session":    response,
// 			},
// 		})
// 		client.Send(*client.ClientId, &messages.Message{
// 			Type: TYPE_START_SESSION,
// 			Body: map[string]interface{}{
// 				"message":    message,
// 				"opponentId": body.OpponentId,
// 				"session":    response,
// 			},
// 		})
// 	}
// }

func (gateway *SessionGatewayImpl) PreparePlayer(client *messages.WebSocketMessager) {
	if client.ClientId == nil {
		client.SendBack(Error(errors.New("not authenticated")))
		return
	}

	var body dto.PreparePlayerRequest
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

	isStarted, err := gateway.SessionService.PreparePlayer(*client.ClientId, body.ChosenHeroList)
	if err != nil {
		client.SendBack(Error(err))
		return
	}

	var message string
	if isStarted {
		session, err := gateway.SessionService.GetSessionByPlayerId(nil, *client.ClientId)
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

		otherId, _ := session.GetOtherPlayerId(*client.ClientId)
		client.Send(otherId, &messages.Message{
			Type: TYPE_START_BATTLE,
			Body: map[string]interface{}{
				"message":    notifyMessage,
				"opponentId": *client.ClientId,
				"session":    response,
			},
		})
		client.Send(*client.ClientId, &messages.Message{
			Type: TYPE_START_BATTLE,
			Body: map[string]interface{}{
				"message":    notifyMessage,
				"opponentId": otherId,
				"session":    response,
			},
		})

		message = "Battle is started.."
		log.Println("Battle is started..")
	} else {
		message = "Waiting for other player.."
	}

	client.SendBack(&messages.Message{
		Type: client.Message.Type,
		Body: map[string]interface{}{
			"success": true,
			"message": message,
		},
	})
}

func (gateway *SessionGatewayImpl) GetServerTime(client *messages.WebSocketMessager) {
	var body dto.ServerTimeRequest
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

	curTime := float64(time.Now().UnixMilli())
	resTime := curTime / 1000.0

	client.SendBack(&messages.Message{
		Type: client.Message.Type,
		Body: map[string]interface{}{
			"localTime":  body.LocalTime,
			"serverTime": resTime,
		},
	})
}

func (gateway *SessionGatewayImpl) NotifySessionCreated(data interface{}) error {
	event, ok := data.(*services.SessionCreatedEvent)
	if !ok {
		log.Println("invalid event class: ")
		panic(data)
	}
	for i, playerId := range event.PlayerIds {
		gateway.Messager.Send(playerId, &messages.Message{
			Type: TYPE_START_SESSION,
			Body: map[string]interface{}{
				"message":    "Session started...",
				"opponentId": event.PlayerIds[(i+1)%2],
				"session":    event.Data,
			},
		})
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
