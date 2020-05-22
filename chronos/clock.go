package chronos

import "time"

var DefaultClock = NewClock(time.Now)

type Clock struct {
	now func() time.Time
}

func NewClock(now func() time.Time) Clock {
	return Clock{now: now}
}

func (c Clock) Now() time.Time {
	return c.now()
}

func (c Clock) Measure(f func() error) (time.Duration, error) {
	then := c.Now()
	err := f()
	return c.Now().Sub(then), err
}
