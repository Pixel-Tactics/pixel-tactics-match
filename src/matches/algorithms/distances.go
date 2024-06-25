package matches_algorithms

import (
	"errors"

	llq "github.com/emirpasic/gods/queues/linkedlistqueue"

	physics "pixeltactics.com/match/src/matches/physics"
)

type BFSElement struct {
	point physics.Point
	dist  int
}

func checkDistanceValidity(mp [][]int, src physics.Point, dest physics.Point) error {
	if len(mp) == 0 || len(mp[0]) == 0 {
		return errors.New("invalid map structure")
	}

	if src.X >= len(mp[0]) || src.Y >= len(mp) {
		return errors.New("invalid src")
	}

	if dest.X >= len(mp[0]) || src.Y >= len(mp) {
		return errors.New("invalid dest")
	}
	return nil
}

func enqueueIfAvailable(queue *llq.Queue, mp [][]int, visited [][]bool, x int, y int, dist int) {
	lenX := len(visited[0])
	lenY := len(visited)

	if !(0 <= x && x < lenX && 0 <= y && y < lenY) {
		return
	}

	if mp[y][x] == 2 {
		return
	}

	if visited[y][x] {
		return
	}

	visited[y][x] = true
	queue.Enqueue(BFSElement{
		point: physics.Point{X: x, Y: y},
		dist:  dist,
	})
}

func CheckDistance(mp [][]int, src physics.Point, dest physics.Point) (int, error) {
	err := checkDistanceValidity(mp, src, dest)
	if err != nil {
		return 0, nil
	}

	lenX := len(mp[0])
	lenY := len(mp)

	visited := make([][]bool, lenY)
	for i := range visited {
		visited[i] = make([]bool, lenX)
	}
	queue := llq.New()
	queue.Enqueue(BFSElement{
		point: src,
		dist:  0,
	})
	visited[src.Y][src.X] = true

	for {
		cur, avail := queue.Dequeue()
		if !avail {
			break
		}

		curElem := cur.(BFSElement)
		curX := curElem.point.X
		curY := curElem.point.Y
		if curElem.point.Equals(dest) {
			return curElem.dist, nil
		} else {
			enqueueIfAvailable(queue, mp, visited, curX+1, curY, curElem.dist+1)
			enqueueIfAvailable(queue, mp, visited, curX-1, curY, curElem.dist+1)
			enqueueIfAvailable(queue, mp, visited, curX, curY+1, curElem.dist+1)
			enqueueIfAvailable(queue, mp, visited, curX, curY-1, curElem.dist+1)
		}
	}
	return 0, errors.New("no path from src to dest")
}
