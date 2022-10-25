package time_util

import (
	"time"
)

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
