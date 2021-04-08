package sources

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitInterval(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, 15*time.Second, timer.waitToNextInterval(timer.startTime))
	})

	t.Run("inside first interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		now := timer.startTime.Add(5 * time.Second)
		assert.Equal(t, 10*time.Second, timer.waitToNextInterval(now))
	})

	t.Run("outside first interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		now := timer.startTime.Add(16 * time.Second)
		assert.Equal(t, 14*time.Second, timer.waitToNextInterval(now))
	})

	t.Run("many intervals late", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		now := timer.startTime.Add(61 * time.Second)
		assert.Equal(t, 14*time.Second, timer.waitToNextInterval(now))
	})

	t.Run("near but still before the interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		now := timer.startTime.Add(14501 * time.Millisecond)
		assert.Equal(t, 15499*time.Millisecond, timer.waitToNextInterval(now))
	})
}
