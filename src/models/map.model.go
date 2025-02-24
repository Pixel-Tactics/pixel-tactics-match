package models

type Map struct {
	SessionId string  `json:"sessionId"`
	Structure [][]int `json:"structure"`
}
