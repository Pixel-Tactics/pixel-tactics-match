package gateway

import (
	"github.com/go-playground/validator/v10"
)

type BaseGateway struct {
	Validator *validator.Validate
}
