package sources

import "time"

type IntervalTimer struct {
	*time.Timer
	interval  time.Duration
	startTime time.Time
}

func (t *IntervalTimer) Reset() {
	waitTime := t.waitToNextInterval(time.Now())
	t.Timer.Reset(waitTime)
}

func NewIntervalTimer(interval time.Duration) *IntervalTimer {
	return &IntervalTimer{
		Timer:     time.NewTimer(interval),
		interval:  interval,
		startTime: time.Now(),
	}
}

func (t *IntervalTimer) waitToNextInterval(now time.Time) time.Duration {
	wait := t.interval - (now.Sub(t.startTime) % t.interval)
	if wait < 500*time.Millisecond {
		wait += t.interval
	}
	return wait
}
