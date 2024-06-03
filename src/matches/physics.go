package matches

const (
	DIRECTION_UP    = "UP"
	DIRECTION_DOWN  = "DOWN"
	DIRECTION_LEFT  = "LEFT"
	DIRECTION_RIGHT = "RIGHT"
)

type Point struct {
	x int
	y int
}

func (p Point) Add(q Point) Point {
	return Point{x: p.x + q.x, y: p.y + q.y}
}

func GetPointFromDirection(dir string) Point {
	if dir == DIRECTION_UP {
		return Point{x: 0, y: -1}
	} else if dir == DIRECTION_DOWN {
		return Point{x: 0, y: 1}
	} else if dir == DIRECTION_LEFT {
		return Point{x: -1, y: 0}
	} else {
		return Point{x: 1, y: 0}
	}
}
