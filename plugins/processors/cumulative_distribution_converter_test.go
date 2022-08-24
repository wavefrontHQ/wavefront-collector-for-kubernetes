package processors

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
)

func TestCumulativeDistributionConversion(t *testing.T) {
	t.Run("Converts cumulative distribution to density distribution", func(t *testing.T) {
		p := NewCumulativeDistributionConverter()

		batch, err := p.Process(&metrics.Batch{Metrics: []wf.Metric{wf.NewCumulativeDistribution(
			"some.distribution",
			"somesource",
			map[string]string{"sometag": "somevalue"},
			[]wf.Centroid{
				{Value: 0.05, Count: 24054},
				{Value: 0.1, Count: 33444},
				{Value: 0.2, Count: 100392},
				{Value: 0.5, Count: 129389},
				{Value: 1, Count: 133988},
				{Value: math.Inf(1), Count: 144320},
			},
			time.Now(),
		)}})

		assert.NoError(t, err)
		assert.Equal(t, 1, len(batch.Metrics))
		distribution := batch.Metrics[0].(*wf.Distribution)
		assert.False(t, distribution.Cumulative)
		assert.Equal(t, 11, len(distribution.Centroids))
	})

	t.Run("Handles conversion failures", func(t *testing.T) {
		p := NewCumulativeDistributionConverter()

		batch, err := p.Process(&metrics.Batch{Metrics: []wf.Metric{wf.NewCumulativeDistribution(
			"some.distribution",
			"somesource",
			map[string]string{"sometag": "somevalue"},
			nil,
			time.Now(),
		)}})

		assert.NoError(t, err)
		assert.Equal(t, 1, len(batch.Metrics))
		distribution := batch.Metrics[0].(*wf.Distribution)
		assert.Equal(t, 0, len(distribution.Centroids))
	})
}
