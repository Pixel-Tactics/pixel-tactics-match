package gateway

import (
	"github.com/go-playground/validator/v10"
)

type BaseGateway struct {
	// Interaction chan *ws_types.Interaction
	// Successes   chan *ws_types.Interaction
	// Errors      chan *ws_types.Interaction
	Validator *validator.Validate
}
