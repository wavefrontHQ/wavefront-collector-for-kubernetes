package processors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
)

func TestDistributionRateCalculator(t *testing.T) {
	t.Run("calculates a rate when it has a previous sample", func(t *testing.T) {
		firstSampleTS := time.Now()
		p := NewDistributionRateCalculator()

		batch, err := p.Process(&metrics.Batch{Metrics: []wf.Metric{wf.NewCumulativeDistribution(
			"some.distribution",
			"somesource",
			map[string]string{"sometag": "somevalue"},
			[]wf.Centroid{{Value: 1, Count: 0}},
			firstSampleTS,
		)}})

		assert.NoError(t, err)
		assert.Equal(t, 0, len(batch.Metrics))

		batch, err = p.Process(&metrics.Batch{Metrics: []wf.Metric{wf.NewCumulativeDistribution(
			"some.distribution",
			"somesource",
			map[string]string{"sometag": "somevalue"},
			[]wf.Centroid{{Value: 1, Count: 1}},
			firstSampleTS.Add(2*time.Minute),
		)}})

		assert.NoError(t, err)
		assert.Equal(t, 1, len(batch.Metrics))

		distribution := batch.Metrics[0].(*wf.Distribution)
		assert.Equal(t, []wf.Centroid{{Value: 1, Count: 0.5}}, distribution.Centroids)
	})

	t.Run("When counter resets metric should not be sent", func(t *testing.T) {
		firstSampleTS := time.Now()
		p := NewDistributionRateCalculator()

		batch, err := p.Process(&metrics.Batch{Metrics: []wf.Metric{
			wf.NewCumulativeDistribution(
				"some.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 10}},
				firstSampleTS,
			),
			wf.NewCumulativeDistribution(
				"another.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 10}},
				firstSampleTS,
			),
		}})

		assert.NoError(t, err)
		assert.Equal(t, 0, len(batch.Metrics))

		batch, err = p.Process(&metrics.Batch{Metrics: []wf.Metric{
			wf.NewCumulativeDistribution(
				"some.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 3}},
				firstSampleTS.Add(time.Minute),
			),
			wf.NewCumulativeDistribution(
				"another.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 3}},
				firstSampleTS.Add(time.Minute),
			),
		}})

		assert.NoError(t, err)
		assert.Equal(t, 0, len(batch.Metrics))

		batch, err = p.Process(&metrics.Batch{Metrics: []wf.Metric{
			wf.NewCumulativeDistribution(
				"some.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 4}},
				firstSampleTS.Add(2*time.Minute),
			),
			wf.NewCumulativeDistribution(
				"another.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 4}},
				firstSampleTS.Add(2*time.Minute),
			),
		}})

		assert.NoError(t, err)
		assert.Equal(t, 2, len(batch.Metrics))
	})

}
