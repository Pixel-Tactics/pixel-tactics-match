package gateway

import (
	"errors"

	"pixeltactics.com/match/src/messages"
)

type Router interface {
	RouteMessage(client *messages.WebSocketMessager)
}

type RouterImpl struct {
	AuthGateway AuthGateway
}

func (router *RouterImpl) RouteMessage(client *messages.WebSocketMessager) {
	if client.Message.Type == "AUTH" {
		router.AuthGateway.AuthenticateClient(client)
	} else {
		client.SendBack(Error(errors.New("invalid type")))
	}
}

func NewRouter(
	authGateway AuthGateway,
) Router {
	return &RouterImpl{
		AuthGateway: authGateway,
	}
}
