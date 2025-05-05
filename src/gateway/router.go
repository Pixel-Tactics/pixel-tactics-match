package gateway

import (
	"errors"

	"pixeltactics.com/match/src/messages"
)

const (
	// Request
	TYPE_INVITE_PLAYER string = "INVITE_PLAYER"
	// TYPE_CREATE_SESSION string = "CREATE_SESSION"
	// TYPE_IS_IN_SESSION  string = "IS_IN_SESSION"
	// TYPE_GET_SESSION    string = "GET_SESSION"

	TYPE_PREPARE_PLAYER string = "PREPARE_PLAYER"
	TYPE_SERVER_TIME    string = "SERVER_TIME"
	TYPE_MOVE           string = "ACTION_MOVE"
	TYPE_ATTACK         string = "ACTION_ATTACK"
	TYPE_END_TURN       string = "ACTION_END_TURN"
	// TYPE_ERROR          string = "ERROR_FEEDBACK"
	// TYPE_FEEDBACK       string = "FEEDBACK"
	// TYPE_ENEMY_ACTION   string = "ENEMY_ACTION"

	// Response
	TYPE_INVITE_SESSION string = "INVITE_SESSION"
	TYPE_START_SESSION  string = "START_SESSION"
	TYPE_START_BATTLE   string = "START_BATTLE"
	TYPE_STATE_CHANGE   string = "STATE_CHANGE"
)

type Router interface {
	RouteMessage(client messages.ClientMessager, message *messages.Message)
}

type RouterImpl struct {
	SessionGateway SessionGateway
	ActionGateway  ActionGateway
}

func (router *RouterImpl) RouteMessage(client messages.ClientMessager, message *messages.Message) {
	switch message.Route {
	case TYPE_PREPARE_PLAYER:
		router.SessionGateway.PreparePlayer(client, message)
	case TYPE_SERVER_TIME:
		router.SessionGateway.GetServerTime(client, message)
	case TYPE_MOVE:
		router.ActionGateway.Move(client, message)
	case TYPE_ATTACK:
		router.ActionGateway.Attack(client, message)
	case TYPE_END_TURN:
		router.ActionGateway.EndTurn(client, message)
	default:
		client.SendBack(Error(errors.New("invalid type")))
	}
}

func NewRouter(
	sessionGateway SessionGateway,
	actionGateway ActionGateway,
) Router {
	return &RouterImpl{
		SessionGateway: sessionGateway,
		ActionGateway:  actionGateway,
	}
}
