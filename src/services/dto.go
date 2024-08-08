package services

import (
	matches_maps "pixeltactics.com/match/src/matches/maps"
)

type CreateSessionRequestDTO struct {
	PlayerId   string `json:"playerId"`
	OpponentId string `json:"opponentId"`
}

type PlayerIdDTO struct {
	PlayerId string `json:"playerId"`
}

type GetSessionResponseDTO struct {
	MatchMap          matches_maps.MatchMap `json:"map"`
	AvailableHeroList []string              `json:"available"`
}

type PreparePlayerRequestDTO struct {
	PlayerId       string   `json:"playerId"`
	ChosenHeroList []string `json:"chosenHeroList"`
}

type PrepareSessionRequestDTO struct {
	PlayerId string
}

type ExecuteActionRequestDTO struct {
	PlayerId       string                 `json:"playerId"`
	ActionName     string                 `json:"actionName"`
	ActionSpecific map[string]interface{} `json:"actionSpecific"`
}
