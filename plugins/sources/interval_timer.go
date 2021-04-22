package sources

import (
	"time"
)

type IntervalTimer struct {
	*time.Timer
	interval      time.Duration
	startTime     time.Time
	lastResetTime time.Time
}

func (t *IntervalTimer) Reset() (intervalsMissed int64) {
	now := time.Now()
	intervals := t.intervalsMissed(now)
	t.lastResetTime = now
	waitTime := t.waitToNextInterval(now.Sub(t.startTime))
	t.Timer.Reset(waitTime)
	return intervals
}

func NewIntervalTimer(interval time.Duration) *IntervalTimer {
	now := time.Now()
	return &IntervalTimer{
		Timer:         time.NewTimer(interval),
		interval:      interval,
		startTime:     now,
		lastResetTime: now,
	}
}

func (t *IntervalTimer) intervalsMissed(now time.Time) (intervalsMissed int64) {
	if now.Sub(t.lastResetTime) < t.interval {
		return 0
	}
	return int64((now.Sub(t.lastResetTime) / t.interval) - 1)
}

func (t *IntervalTimer) waitToNextInterval(diff time.Duration) time.Duration {
	wait := t.interval - (diff % t.interval)
	if wait < scaleInterval(t.interval, 0.0333) { // 3.33%. This was chosen arbitrarily. If you have a better idea, change it!
		wait += t.interval
	}
	return wait
}

func scaleInterval(interval time.Duration, ratio float64) time.Duration {
	return time.Duration(float64(interval) * ratio)
}
