package services

import "errors"

var ErrEmptyHero = errors.New("player doesn't have hero")
var ErrInvalidPlayer = errors.New("invalid player")
