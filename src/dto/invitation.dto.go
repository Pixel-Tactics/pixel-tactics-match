package dto

type InvitationRequest struct {
	Id         string `json:"id" validate:"required,min=1"`
	PlayerId   string `json:"sourceUsername" validate:"required,min=1"`
	OpponentId string `json:"destinationUsername" validate:"required,min=1"`
}
