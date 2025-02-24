package gateway

import (
	"errors"
	"log"

	"pixeltactics.com/match/src/messages"
)

const (
	TYPE_CREATE_SESSION string = "CREATE_SESSION"
	TYPE_IS_IN_SESSION  string = "IS_IN_SESSION"
	TYPE_GET_SESSION    string = "GET_SESSION"
	TYPE_INVITE_SESSION string = "INVITE_SESSION"
	TYPE_START_SESSION  string = "START_SESSION"
	TYPE_PREPARE_PLAYER string = "PREPARE_PLAYER"
	TYPE_SERVER_TIME    string = "SERVER_TIME"
	TYPE_AUTH           string = "AUTH"
	TYPE_ERROR          string = "ERROR_FEEDBACK"
	TYPE_FEEDBACK       string = "FEEDBACK"
	TYPE_ENEMY_ACTION   string = "ENEMY_ACTION"
	TYPE_START_BATTLE   string = "START_BATTLE"
	TYPE_EXECUTE_ACTION string = "EXECUTE_ACTION"
	TYPE_APPLY_ACTION   string = "APPLY_ACTION"
	TYPE_END_TURN       string = "END_TURN"
)

type Router interface {
	RouteMessage(client *messages.WebSocketMessager)
}

type RouterImpl struct {
	AuthGateway    AuthGateway
	SessionGateway SessionGateway
}

func (router *RouterImpl) RouteMessage(client *messages.WebSocketMessager) {
	log.Println("TOOO SHINJINTERU")
	if client.Message.Type == TYPE_AUTH {
		router.AuthGateway.AuthenticateClient(client)
	} else if client.Message.Type == TYPE_CREATE_SESSION {
		log.Println("HERE")
		router.SessionGateway.CreateSession(client)
	} else {
		client.SendBack(Error(errors.New("invalid type")))
	}
}

func NewRouter(
	authGateway AuthGateway,
	sessionGateway SessionGateway,
) Router {
	return &RouterImpl{
		AuthGateway:    authGateway,
		SessionGateway: sessionGateway,
	}
}
