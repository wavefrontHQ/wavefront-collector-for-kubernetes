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

	t.Run("handles duplicate series", func(t *testing.T) {
		firstSampleTS := time.Now()
		p := NewDistributionRateCalculator()
		DuplicateHistogramCounter("some.distribution").Clear()

		batch, err := p.Process(&metrics.Batch{Metrics: []wf.Metric{
			wf.NewCumulativeDistribution(
				"some.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 0}},
				firstSampleTS,
			),
			wf.NewCumulativeDistribution(
				"some.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 1}},
				firstSampleTS.Add(time.Second),
			),
		}})

		assert.NoError(t, err)
		assert.Equal(t, 0, len(batch.Metrics))
		assert.Equal(t, int64(1), DuplicateHistogramCounter("some.distribution").Count())

		batch, err = p.Process(&metrics.Batch{Metrics: []wf.Metric{
			wf.NewCumulativeDistribution(
				"some.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 1}},
				firstSampleTS.Add(time.Minute),
			),
			wf.NewCumulativeDistribution(
				"some.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 2}},
				firstSampleTS.Add(time.Minute+time.Second),
			),
		}})

		assert.NoError(t, err)
		assert.Equal(t, 1, len(batch.Metrics))
		assert.Equal(t, int64(2), DuplicateHistogramCounter("some.distribution").Count())
	})

	t.Run("can mutate distributions after processing without affecting rate calculation", func(t *testing.T) {
		firstSampleTS := time.Now()
		p := NewDistributionRateCalculator()
		DuplicateHistogramCounter("some.distribution").Clear()

		firstDistribution := wf.NewCumulativeDistribution(
			"some.distribution",
			"somesource",
			map[string]string{"sometag": "somevalue"},
			[]wf.Centroid{{Value: 1, Count: 0}},
			firstSampleTS,
		)
		batch, _ := p.Process(&metrics.Batch{Metrics: []wf.Metric{firstDistribution}})

		firstDistribution.AddTags(map[string]string{"extra": "foo"})

		batch, _ = p.Process(&metrics.Batch{Metrics: []wf.Metric{
			wf.NewCumulativeDistribution(
				"some.distribution",
				"somesource",
				map[string]string{"sometag": "somevalue"},
				[]wf.Centroid{{Value: 1, Count: 1}},
				firstSampleTS.Add(time.Minute),
			),
		}})

		assert.Equal(t, 1, len(batch.Metrics))
	})
}
