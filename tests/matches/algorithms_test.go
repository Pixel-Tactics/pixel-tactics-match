package matches_test

import (
	"testing"

	matches_algorithms "pixeltactics.com/match/src/matches/algorithms"
	matches_physics "pixeltactics.com/match/src/matches/physics"
)

func TestCheckDistance(t *testing.T) {
	mp := [][]int{
		{1, 1, 1, 1},
		{2, 3, 1, 1},
		{2, 2, 1, 1},
		{2, 2, 1, 1},
		{1, 1, 1, 1},
	}
	dist, err := matches_algorithms.CheckDistance(mp, matches_physics.Point{X: 0, Y: 4}, matches_physics.Point{X: 1, Y: 1})
	if err != nil {
		t.Error("expected no error", "got", err)
	}
	if dist != 6 {
		t.Error("expected", 6, "got", dist)
	}
}
