package sources

import "time"

type IntervalTimer struct {
	*time.Timer
	interval  time.Duration
	startTime time.Time
}

func (t *IntervalTimer) Reset() int64 {
	diff := time.Now().Sub(t.startTime)
	waitTime := t.waitToNextInterval(diff)
	t.Timer.Reset(waitTime)
	return int64((diff + waitTime) / t.interval)
}

func NewIntervalTimer(interval time.Duration) *IntervalTimer {
	return &IntervalTimer{
		Timer:     time.NewTimer(interval),
		interval:  interval,
		startTime: time.Now(),
	}
}

func (t *IntervalTimer) waitToNextInterval(diff time.Duration) time.Duration {
	wait := t.interval - (diff % t.interval)
	per10K := 333 // 3.33%
	if wait < scaleInterval(t.interval, per10K) {
		wait += t.interval
	}
	return wait
}

func scaleInterval(interval time.Duration, per10K int) time.Duration {
	return (interval*time.Duration(per10K) + 10_000 - 1) / 10_000
}
