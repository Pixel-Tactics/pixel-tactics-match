package dto

type Auth struct {
	PlayerToken string `json:"playerToken" validate:"required,min=1"`
}
