package sources

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// same time
// inside interval
// outside interval
// +/- near interval
func TestWaitInterval(t *testing.T) {
	t.Run("zero time", func(t *testing.T) {
		interval := 15 * time.Second
		period := 0 * time.Second
		assert.Equal(t, 15*time.Second, waitInterval(interval, period))
	})

	t.Run("inside first interval", func(t *testing.T) {
		interval := 15 * time.Second
		period := 5 * time.Second
		assert.Equal(t, 10*time.Second, waitInterval(interval, period))
	})

	t.Run("outside first interval", func(t *testing.T) {
		interval := 15 * time.Second
		period := 16 * time.Second
		assert.Equal(t, 14*time.Second, waitInterval(interval, period))
	})

	t.Run("many intervals late", func(t *testing.T) {
		interval := 15 * time.Second
		period := 61 * time.Second
		assert.Equal(t, 14*time.Second, waitInterval(interval, period))
	})

	t.Run("near but still before the interval", func(t *testing.T) {
		interval := 15 * time.Second
		period := 14501 * time.Millisecond
		assert.Equal(t, 15499*time.Millisecond, waitInterval(interval, period))
	})
}
