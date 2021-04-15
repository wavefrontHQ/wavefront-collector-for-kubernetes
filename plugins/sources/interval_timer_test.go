package sources

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitInterval(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, 15*time.Second, timer.waitToNextInterval(0))
	})

	t.Run("inside first interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, 10*time.Second, timer.waitToNextInterval(5*time.Second))
	})

	t.Run("outside first interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, 14*time.Second, timer.waitToNextInterval(16*time.Second))
	})

	t.Run("many intervals late", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, 14*time.Second, timer.waitToNextInterval(61*time.Second))
	})

	t.Run("near but still before the interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, 15499*time.Millisecond, timer.waitToNextInterval(14501*time.Millisecond))
	})
}

func TestReset(t *testing.T) {
	t.Run("lastResetTime is set", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		timer.Reset()
		//assert.Equal(t, time.Now(),timer.lastResetTime)
		assert.NotZero(t, timer.lastResetTime)
	})

	t.Run("timer is reset", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Millisecond)
		calls := make(chan int, 0)
		missed := make(chan int64, 0)
		go func(calls chan int, missed chan int64) {
			calledCurrent, missedCurrent := 0, int64(0)
			for {
				select {
				case <-timer.C:
					fmt.Println("in timer call")
					calledCurrent = <-calls
					missedCurrent = <-missed
					calls <- calledCurrent + 1
					missed <- timer.intervalsMissed() + missedCurrent
					timer.Reset()
				}
			}
		}(calls, missed)
		calls <- 0
		missed <- 0
		time.Sleep(3 * time.Second)
		assert.Equal(t, 1, <-calls, "timer wasn't called the expected number of times")
		assert.Zero(t, <-missed)
	})
}

func TestMissedCount(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, int64(0), timer.intervalsMissed())
	})

	t.Run("first interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		timer.lastResetTime = time.Now().Add(-15 * time.Second)
		assert.Equal(t, int64(0), timer.intervalsMissed())
	})

	t.Run("outside second interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		timer.lastResetTime = time.Now().Add(-31 * time.Second)
		assert.Equal(t, int64(1), timer.intervalsMissed())
	})

	t.Run("many intervals late", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		timer.lastResetTime = time.Now().Add(-61 * time.Second)
		assert.Equal(t, int64(3), timer.intervalsMissed())
	})

	t.Run("near but still before the interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		timer.lastResetTime = time.Now().Add(-14501 * time.Millisecond)
		assert.Equal(t, int64(0), timer.intervalsMissed())
	})

}
