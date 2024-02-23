package time_util

import (
	"time"
)

const DateTimeFormat = "2006-01-02 15:04:05"

func MaxTime(first, second time.Duration) time.Duration {
	if first > second {
		return first
	}
	return second
}

func MinTime(first, second time.Duration) time.Duration {
	if first < second {
		return first
	}
	return second
}
