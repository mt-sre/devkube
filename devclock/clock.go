package devclock

import (
	"time"

	"k8s.io/utils/clock"
)

type RealClock = clock.RealClock

type Clock interface {
	NewTicker(d time.Duration) clock.Ticker
}

func RealClockIfUnset(c Clock) Clock {
	if c == nil {
		return clock.RealClock{}
	}
	return c
}
