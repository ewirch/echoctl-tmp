package schedule

import (
	"time"
)

// timer is an abstraction around clock.Timer, so we can use different implementations in code. F.e.: foreverTimer.
type timer interface {
	c() <-chan time.Time
	stop()
}

type foreverTimer struct {
	ch <-chan time.Time
}

func (f *foreverTimer) c() <-chan time.Time {
	return f.ch
}

func (f *foreverTimer) stop() {
	// nop
}

func newForEverTimer() timer {
	// This channel never gets input, or is closed, therefore, waiting on ch will wait forever.
	timer := foreverTimer{
		ch: make(chan time.Time),
	}
	return &timer
}

type clockTimer struct {
	*time.Timer
}

func (c clockTimer) c() <-chan time.Time {
	return c.C
}

func (c clockTimer) stop() {
	c.Stop()
}

func newTimeTimer(triggerAt time.Time) timer {
	return &clockTimer{
		time.NewTimer(time.Until(triggerAt)),
	}
}
