package sources

import (
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
		assert.NotZero(t, timer.lastResetTime)
	})
}

func TestMissedCount(t *testing.T) {
	now := time.Now()
	t.Run("no resets", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, int64(0), timer.intervalsMissed(time.Now()))
	})

	t.Run("first interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		timer.lastResetTime = now.Add(-15 * time.Second)
		assert.Equal(t, int64(0), timer.intervalsMissed(now))
	})

	t.Run("outside second interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		timer.lastResetTime = now.Add(-31 * time.Second)
		assert.Equal(t, int64(1), timer.intervalsMissed(now))
	})

	t.Run("many intervals late", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		timer.lastResetTime = now.Add(-61 * time.Second)
		assert.Equal(t, int64(3), timer.intervalsMissed(now))
	})

	t.Run("near but still before the interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		timer.lastResetTime = now.Add(-14501 * time.Millisecond)
		assert.Equal(t, int64(0), timer.intervalsMissed(now))
	})
}
