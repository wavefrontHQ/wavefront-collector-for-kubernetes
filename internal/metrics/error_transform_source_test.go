package metrics

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type dummyMetricsSource struct {
	name string
	cleanupCalled bool
	dataBatch *DataBatch
}

func (d *dummyMetricsSource) Name() string {
	return d.name
}

func (d *dummyMetricsSource) ScrapeMetrics() (*DataBatch, error) {
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
		src := NewErrorTransformSource(&dummyMetricsSource{name:"name"}, func(err error) error{return errors.New("error")})
		assert.Equal(t, "name", src.Name())
	})

	t.Run("cleans up the inner source", func(t *testing.T) {
		d := &dummyMetricsSource{name:"name"}
		src := NewErrorTransformSource(d, func(err error) error{return errors.New("error")})
		src.Cleanup()
		d.VerifyCleanupCalled(t)
	})

	t.Run("transforms the error when scraping metrics", func(t *testing.T) {
		d := &dummyMetricsSource{name:"name"}
		src := NewErrorTransformSource(d, func(err error) error{return errors.New("custom error")})
		_, err := src.ScrapeMetrics()
		assert.Equal(t, "custom error", err.Error())
	})

	t.Run("preserves the DataBatch when scraping metrics", func(t *testing.T) {
		expectedDataBatch := &DataBatch{Timestamp: time.Now()}
		d := &dummyMetricsSource{name:"name", dataBatch: expectedDataBatch}
		src := NewErrorTransformSource(d, func(err error) error{return errors.New("custom error")})
		actualDataBatch, _ := src.ScrapeMetrics()
		assert.Equal(t, expectedDataBatch, actualDataBatch)
	})
}

