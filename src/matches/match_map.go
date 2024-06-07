package matches

// 0 = Background
// 1 = Land
// 2 = Obstacle
// 3 = Player 1 Spawn
// 4 = Player 2 Spawn
type MatchMap struct {
	Structure [][]int `json:"structure"`
}

func (m *MatchMap) IsPointOpen(pos Point) bool {
	curValue := m.Structure[pos.Y][pos.X]
	return curValue != 2
}

func GenerateMap() (*MatchMap, error) {
	structure := [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 2, 2, 2, 2, 2, 2},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 4, 3, 1, 1, 1, 1},
		{2, 2, 2, 2, 2, 2, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	}
	newMap := MatchMap{
		Structure: structure,
	}
	return &newMap, nil
}
