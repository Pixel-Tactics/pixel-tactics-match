package services

import "errors"

var ErrEmptyHero = errors.New("player doesn't have hero")
var ErrInvalidPlayer = errors.New("invalid player")

var ErrNotTurn = errors.New("this turn is for other player")
var ErrSessionNotFound = errors.New("session not found")
var ErrInSession = errors.New("player is in session")
