package services

import "pixeltactics.com/match/src/matches"

type CreateSessionRequestDTO struct {
	PlayerId   string `json:"playerId"`
	OpponentId string `json:"opponentId"`
}

type GetSessionRequestDTO struct {
	SessionId string `json:"sessionId"`
}

type GetSessionResponseDTO struct {
	MatchMap          matches.MatchMap `json:"map"`
	AvailableHeroList []string         `json:"available"`
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
