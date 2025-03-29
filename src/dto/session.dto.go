package dto

import (
	matches_maps "pixeltactics.com/match/src/matches/maps"
)

type CreateSessionRequest struct {
	// PlayerId   string `json:"playerId" validate:"required,min=1"`
	OpponentId string `json:"opponentId" validate:"required,min=1"`
}

type PlayerId struct {
	PlayerId string `json:"playerId" validate:"required,min=1"`
}

type GetSessionResponse struct {
	MatchMap          matches_maps.MatchMap `json:"map"`
	AvailableHeroList []string              `json:"available"`
}

type PreparePlayerRequest struct {
	// PlayerId       string   `json:"playerId"`
	ChosenHeroList []string `json:"chosenHeroList"`
}

type PrepareSessionRequest struct {
	PlayerId string
}

type ExecuteActionRequest struct {
	PlayerId       string                 `json:"playerId"`
	ActionName     string                 `json:"actionName"`
	ActionSpecific map[string]interface{} `json:"actionSpecific"`
}
