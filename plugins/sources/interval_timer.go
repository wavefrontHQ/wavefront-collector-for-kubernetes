package sources

import "time"

type IntervalTimer struct {
	*time.Timer
	interval      time.Duration
	startTime     time.Time
	lastResetTime time.Time
}

func (t *IntervalTimer) Reset() {
	nowTime := time.Now()
	t.lastResetTime = nowTime
	diff := nowTime.Sub(t.startTime)
	waitTime := t.waitToNextInterval(diff)
	t.Timer.Reset(waitTime)
}

func (t *IntervalTimer) intervalsMissed() (intervalsMissed int64) {
	nowTime := time.Now()
	if t.lastResetTime.IsZero() || nowTime.Sub(t.lastResetTime) < t.interval {
		return 0
	}
	return int64((nowTime.Sub(t.lastResetTime) / t.interval) - 1)
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
	per10K := 333 // 3.33%. This was chosen arbitrarily. If you have a better idea, change it!
	if wait < scaleInterval(t.interval, per10K) {
		wait += t.interval
	}
	return wait
}

func scaleInterval(interval time.Duration, per10K int) time.Duration {
	return (interval*time.Duration(per10K) + 10_000 - 1) / 10_000
}
