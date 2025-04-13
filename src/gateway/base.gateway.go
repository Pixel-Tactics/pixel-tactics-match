package gateway

import (
	"github.com/go-playground/validator/v10"
	"pixeltactics.com/match/src/messages"
)

type BaseGateway struct {
	Messager  messages.Messager
	Validator *validator.Validate
}
