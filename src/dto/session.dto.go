package dto

type CreateSessionRequest struct {
	OpponentId string `json:"opponentId" validate:"required,min=1"`
}

type PlayerId struct {
	PlayerId string `json:"playerId" validate:"required,min=1"`
}

type PreparePlayerRequest struct {
	ChosenHeroList []string `json:"chosenHeroList" validate:"required,min=1"`
}

type ServerTimeRequest struct {
	LocalTime float64 `json:"localTime" validate:"required"`
}
