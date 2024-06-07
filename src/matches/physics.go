package matches

const (
	DIRECTION_UP    = "UP"
	DIRECTION_DOWN  = "DOWN"
	DIRECTION_LEFT  = "LEFT"
	DIRECTION_RIGHT = "RIGHT"
)

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (p Point) Add(q Point) Point {
	return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

func (p Point) Equals(o Point) bool {
	return p.X == o.X && p.Y == o.Y
}

func GetPointFromDirection(dir string) Point {
	if dir == DIRECTION_UP {
		return Point{X: 0, Y: -1}
	} else if dir == DIRECTION_DOWN {
		return Point{X: 0, Y: 1}
	} else if dir == DIRECTION_LEFT {
		return Point{X: -1, Y: 0}
	} else {
		return Point{X: 1, Y: 0}
	}
}
