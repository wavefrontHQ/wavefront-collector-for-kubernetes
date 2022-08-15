package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
)

func TestBatch(t *testing.T) {
	t.Run("Points", func(t *testing.T) {
		t.Run("counts Metrics", func(t *testing.T) {
			b := &metrics.Batch{
				Metrics: []wf.Metric{
					wf.NewPoint("some.point", 50.0, 0.0, "somepointsource", map[string]string{}),
					wf.NewFrequencyDistribution("some.distribution", "somedistrosource", map[string]string{}, []wf.Centroid{}, time.Now()),
				},
			}

			assert.Equal(t, 8, b.Points())
		})

		t.Run("counts Set.Values", func(t *testing.T) {
			b := &metrics.Batch{
				Sets: map[metrics.ResourceKey]*metrics.Set{
					metrics.PodKey("somenamespace", "somepod"): {
						Values: map[string]metrics.Value{
							"some.metric":    {},
							"another.metric": {},
						},
					},
					metrics.PodKey("somenamespace", "anotherpod"): {
						Values: map[string]metrics.Value{
							"some.metric":    {},
							"another.metric": {},
						},
					},
				},
			}

			assert.Equal(t, 4, b.Points())
		})

		t.Run("counts Set.LabeledValues", func(t *testing.T) {
			b := &metrics.Batch{
				Sets: map[metrics.ResourceKey]*metrics.Set{
					metrics.PodKey("somenamespace", "somepod"): {
						LabeledValues: []metrics.LabeledValue{{Name: "some.metric"}, {Name: "another.metric"}},
					},
					metrics.PodKey("somenamespace", "anotherpod"): {
						LabeledValues: []metrics.LabeledValue{{Name: "some.metric"}, {Name: "another.metric"}},
					},
				},
			}

			assert.Equal(t, 4, b.Points())
		})
	})
}
