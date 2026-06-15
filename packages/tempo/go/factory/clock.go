package factory

import "time"

type Clock struct {
	now *time.Time
}

func NewClock(now *time.Time) Clock {
	return Clock{now: now}
}

func (clock Clock) Now() time.Time {
	if clock.now != nil {
		return *clock.now
	}

	return time.Now()
}
