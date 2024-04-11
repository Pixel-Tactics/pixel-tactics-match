package matches

import (
	"errors"
)

type Player struct {
	Id     string `json:"id"`
	Secret string `json:"secret"`
}

type MatchMap struct {
	Structure [][]int `json:"structure"`
}

func GenerateMap() (*MatchMap, error) {
	structure := [][]int{
		{1, 1, 1, 1, 1},
		{1, 2, 2, 1, 1},
		{1, 1, 1, 1, 1},
		{1, 1, 2, 2, 1},
		{1, 1, 1, 1, 1},
	}
	newMap := MatchMap{
		Structure: structure,
	}
	return &newMap, nil
}

type Hero struct {
	Health    int
	MaxHealth int
	Damage    int
}

func GetAvailableHeroes() (*[]string, error) {
	return &[]string{"knight"}, nil
}

type Session struct {
	Id                string
	Secret            string
	Player1           *Player
	Player2           *Player
	Running           bool
	MatchMap          *MatchMap
	AvailableHeroList *[]string
}

func (session Session) GetOpponentPlayer(player *Player) (*Player, error) {
	if player.Id == session.Player1.Id {
		return session.Player2, nil
	} else if player.Id == session.Player2.Id {
		return session.Player1, nil
	} else {
		return nil, errors.New("player in session")
	}
}
