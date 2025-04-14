package dto

type InvitationRequest struct {
	PlayerId string `json:"opponentId" validate:"required,min=1"`
}
