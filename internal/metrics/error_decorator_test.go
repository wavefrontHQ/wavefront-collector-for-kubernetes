package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type dummyMetricsSource struct {
	autoDiscovered bool
	name           string
	cleanupCalled  bool
	dataBatch      *Batch
}

func (d *dummyMetricsSource) AutoDiscovered() bool {
	return d.autoDiscovered
}

func (d *dummyMetricsSource) Name() string {
	return d.name
}

func (d *dummyMetricsSource) Scrape() (*Batch, error) {
	return d.dataBatch, nil
}

func (d *dummyMetricsSource) Cleanup() {
	d.cleanupCalled = true
}

func (d *dummyMetricsSource) VerifyCleanupCalled(t *testing.T) {
	assert.True(t, d.cleanupCalled)
}

func TestErrorTransformSource(t *testing.T) {
	t.Run("takes on the name of the inner source", func(t *testing.T) {
		src := NewErrorDecorator(&dummyMetricsSource{name: "name"}, func(err error) error { return errors.New("error") })
		assert.Equal(t, "name", src.Name())
	})

	t.Run("cleans up the inner source", func(t *testing.T) {
		d := &dummyMetricsSource{name: "name"}
		src := NewErrorDecorator(d, func(err error) error { return errors.New("error") })
		src.Cleanup()
		d.VerifyCleanupCalled(t)
	})

	t.Run("transforms the error when scraping metrics", func(t *testing.T) {
		d := &dummyMetricsSource{name: "name"}
		src := NewErrorDecorator(d, func(err error) error { return errors.New("custom error") })
		_, err := src.Scrape()
		assert.Equal(t, "custom error", err.Error())
	})

	t.Run("preserves the Batch when scraping metrics", func(t *testing.T) {
		expectedDataBatch := &Batch{Timestamp: time.Now()}
		d := &dummyMetricsSource{name: "name", dataBatch: expectedDataBatch}
		src := NewErrorDecorator(d, func(err error) error { return errors.New("custom error") })
		actualDataBatch, _ := src.Scrape()
		assert.Equal(t, expectedDataBatch, actualDataBatch)
	})
}
