package utils

import "time"

func MinTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
