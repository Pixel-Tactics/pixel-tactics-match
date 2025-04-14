package gateway

import (
	"github.com/go-playground/validator/v10"
	"pixeltactics.com/match/src/messages"
)

type HasMessager interface {
	SetMessager(messager messages.Messager)
}

type BaseGateway struct {
	Messager  messages.Messager
	Validator *validator.Validate
}

func (gateway *BaseGateway) SetMessager(messager messages.Messager) {
	gateway.Messager = messager
}
