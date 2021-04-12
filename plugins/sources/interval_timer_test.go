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
		assert.Equal(t, 10*time.Second, timer.waitToNextInterval(5 * time.Second))
	})

	t.Run("outside first interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, 14*time.Second, timer.waitToNextInterval(16 * time.Second))
	})

	t.Run("many intervals late", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, 14*time.Second, timer.waitToNextInterval(61 * time.Second))
	})

	t.Run("near but still before the interval", func(t *testing.T) {
		timer := NewIntervalTimer(15 * time.Second)
		assert.Equal(t, 15499*time.Millisecond, timer.waitToNextInterval(14501 * time.Millisecond))
	})
}

func TestMissedCount(t *testing.T) {
	t.Log("TODO: write test to drive out missed count")
	t.Fail()
}
