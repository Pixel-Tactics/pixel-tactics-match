package matches_test

import (
	"testing"

	"pixeltactics.com/match/src/matches"
)

func TestCheckDistance(t *testing.T) {
	mp := [][]int{
		{1, 1, 1, 1},
		{2, 3, 1, 1},
		{2, 2, 1, 1},
		{2, 2, 1, 1},
		{1, 1, 1, 1},
	}
	dist, err := matches.CheckDistance(mp, matches.Point{X: 0, Y: 4}, matches.Point{X: 1, Y: 1})
	if err != nil {
		t.Error("expected no error", "got", err)
	}
	if dist != 6 {
		t.Error("expected", 6, "got", dist)
	}
}
