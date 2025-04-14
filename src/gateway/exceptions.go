package gateway

import "errors"

var ErrNotAuthenticated = errors.New("not authenticated")
var ErrInvalidInput = errors.New("invalid input")
