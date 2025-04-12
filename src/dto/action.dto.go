package dto

import "pixeltactics.com/match/src/utils/physics"

type MoveRequest struct {
	Hero       string              `json:"hero" validate:"required,min=1"`
	Directions []physics.Direction `json:"directions" validate:"required,min=1,dive,oneof=UP DOWN LEFT RIGHT"`
}

type AttackRequest struct {
	SrcHero string `json:"srcHero" validate:"required,min=1"`
	DstHero string `json:"dstHero" validate:"required,min=1"`
}
