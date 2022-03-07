package processors

import (
	"testing"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

	"github.com/stretchr/testify/assert"
)

func TestSumCountAggregator(t *testing.T) {
	t.Run("has a customizable name", func(t *testing.T) {
		sca := NewSumCountAggregator("my_resource", []SumCountAggregateSpec{})
		assert.Equal(t, sca.Name(), "my_resource_aggregator")
	})

	t.Run("adds a group set when the group can be extracted", func(t *testing.T) {
		groupKey := metrics.ResourceKey("group")
		sca := NewSumCountAggregator("my_resource", []SumCountAggregateSpec{{
			ResourceSumMetrics: []string{},
			IsPartOfGroup: func(_ *metrics.Set) bool {
				return true
			},
			Group: func(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set) {
				return groupKey, &metrics.Set{
					Labels: map[string]string{},
					Values: map[string]metrics.Value{},
				}
			},
		}})
		inputBatch := &metrics.Batch{
			Timestamp: time.Now(),
			Sets: map[metrics.ResourceKey]*metrics.Set{
				metrics.PodKey("ns1", "pod1"): {
					Labels: map[string]string{},
					Values: map[string]metrics.Value{},
				},
			},
		}

		outputBatch, err := sca.Process(inputBatch)
		assert.NoError(t, err)

		groupSet := outputBatch.Sets[groupKey]
		assert.NotNil(t, groupSet)
	})

	t.Run("does not add a group set the group can't be extracted", func(t *testing.T) {
		groupKey := metrics.ResourceKey("group")
		sca := NewSumCountAggregator("my_resource", []SumCountAggregateSpec{{
			ResourceSumMetrics: []string{},
			IsPartOfGroup: func(_ *metrics.Set) bool {
				return true
			},
			Group: func(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set) {
				return groupKey, nil
			},
		}})
		inputBatch := &metrics.Batch{
			Timestamp: time.Now(),
			Sets: map[metrics.ResourceKey]*metrics.Set{
				metrics.PodKey("ns1", "pod1"): {
					Labels: map[string]string{},
					Values: map[string]metrics.Value{},
				},
			},
		}

		outputBatch, err := sca.Process(inputBatch)
		assert.NoError(t, err)

		_, found := outputBatch.Sets[groupKey]
		assert.False(t, found, "does not add groupKey to the outputBatch")
	})

	t.Run("sums metrics by group", func(t *testing.T) {
		groupKey := metrics.ResourceKey("group")
		sca := NewSumCountAggregator("my_resource", []SumCountAggregateSpec{{
			ResourceSumMetrics: []string{"m1"},
			IsPartOfGroup: func(resourceSet *metrics.Set) bool {
				return resourceSet.Labels["type"] == "pod"
			},
			Group: func(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set) {
				groupSet := batch.Sets[groupKey]
				if groupSet == nil {
					groupSet = &metrics.Set{
						Labels: map[string]string{},
						Values: map[string]metrics.Value{},
					}
				}
				return groupKey, groupSet
			},
		}})
		inputBatch := &metrics.Batch{
			Timestamp: time.Now(),
			Sets: map[metrics.ResourceKey]*metrics.Set{
				metrics.PodKey("ns1", "pod1"): {
					Labels: map[string]string{"type": "pod"},
					Values: map[string]metrics.Value{
						"m1": {
							ValueType: metrics.ValueInt64,
							IntValue:  1,
						},
						"m2": {
							ValueType: metrics.ValueInt64,
							IntValue:  2,
						},
					},
				},
				metrics.PodKey("ns1", "pod2"): {
					Labels: map[string]string{"type": "pod"},
					Values: map[string]metrics.Value{"m1": {
						ValueType: metrics.ValueInt64,
						IntValue:  3,
					}},
				},
			},
		}

		outputBatch, err := sca.Process(inputBatch)
		assert.NoError(t, err)

		groupSet := outputBatch.Sets[groupKey]
		assert.NotNil(t, groupSet, "groupSet should exist")

		assert.Equal(t, int64(4), groupSet.Values["m1"].IntValue)
	})

	t.Run("counts resources by group", func(t *testing.T) {
		groupKey := metrics.ResourceKey("group")
		sca := NewSumCountAggregator("my_resource", []SumCountAggregateSpec{{
			ResourceSumMetrics:  []string{"m1"},
			ResourceCountMetric: "c1",
			IsPartOfGroup: func(resourceSet *metrics.Set) bool {
				return resourceSet.Labels["type"] == "pod"
			},
			Group: func(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set) {
				groupSet := batch.Sets[groupKey]
				if groupSet == nil {
					groupSet = &metrics.Set{
						Labels: map[string]string{},
						Values: map[string]metrics.Value{},
					}
				}
				return groupKey, groupSet
			},
		}})
		inputBatch := &metrics.Batch{
			Timestamp: time.Now(),
			Sets: map[metrics.ResourceKey]*metrics.Set{
				metrics.PodKey("ns1", "pod1"): {
					Labels: map[string]string{"type": "pod"},
					Values: map[string]metrics.Value{"m1": {
						ValueType: metrics.ValueInt64,
						IntValue:  1,
					}},
				},
				metrics.PodKey("ns1", "pod2"): {
					Labels: map[string]string{"type": "pod"},
					Values: map[string]metrics.Value{"m1": {
						ValueType: metrics.ValueInt64,
						IntValue:  2,
					}},
				},
			},
		}

		outputBatch, err := sca.Process(inputBatch)
		assert.NoError(t, err)

		groupSet := outputBatch.Sets[groupKey]
		assert.NotNil(t, groupSet, "groupSet should exist")

		assert.Equal(t, int64(2), groupSet.Values["c1"].IntValue)
	})

	t.Run("sums existing resource counts by group", func(t *testing.T) {
		groupKey := metrics.ResourceKey("group")
		sca := NewSumCountAggregator("my_resource", []SumCountAggregateSpec{{
			ResourceSumMetrics:  []string{"m1"},
			ResourceCountMetric: "c1",
			IsPartOfGroup: func(resourceSet *metrics.Set) bool {
				return resourceSet.Labels["type"] == "pod"
			},
			Group: func(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set) {
				groupSet := batch.Sets[groupKey]
				if groupSet == nil {
					groupSet = &metrics.Set{
						Labels: map[string]string{},
						Values: map[string]metrics.Value{},
					}
				}
				return groupKey, groupSet
			},
		}})
		inputBatch := &metrics.Batch{
			Timestamp: time.Now(),
			Sets: map[metrics.ResourceKey]*metrics.Set{
				metrics.PodKey("ns1", "pod1"): {
					Labels: map[string]string{"type": "pod"},
					Values: map[string]metrics.Value{
						"m1": {
							ValueType: metrics.ValueInt64,
							IntValue:  1,
						},
						"c1": {
							ValueType: metrics.ValueInt64,
							IntValue:  2,
						},
					},
				},
				metrics.PodKey("ns1", "pod2"): {
					Labels: map[string]string{"type": "pod"},
					Values: map[string]metrics.Value{
						"m1": {
							ValueType: metrics.ValueInt64,
							IntValue:  3,
						},
						"c1": {
							ValueType: metrics.ValueInt64,
							IntValue:  4,
						},
					},
				},
			},
		}

		outputBatch, err := sca.Process(inputBatch)
		assert.NoError(t, err)

		groupSet := outputBatch.Sets[groupKey]
		assert.NotNil(t, groupSet, "groupSet should exist")

		assert.Equal(t, int64(6), groupSet.Values["c1"].IntValue)
	})

	t.Run("handles multiple specs", func(t *testing.T) {
		groupKey := metrics.ResourceKey("group")
		sca := NewSumCountAggregator("my_resource", []SumCountAggregateSpec{
			{
				ResourceSumMetrics:  []string{"m1"},
				ResourceCountMetric: "c1",
				IsPartOfGroup: func(resourceSet *metrics.Set) bool {
					return resourceSet.Labels["type"] == "pod"
				},
				Group: func(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set) {
					groupSet := batch.Sets[groupKey]
					if groupSet == nil {
						groupSet = &metrics.Set{
							Labels: map[string]string{},
							Values: map[string]metrics.Value{},
						}
					}
					return groupKey, groupSet
				},
			},
			{
				ResourceSumMetrics:  []string{"m2"},
				ResourceCountMetric: "c2",
				IsPartOfGroup: func(resourceSet *metrics.Set) bool {
					return resourceSet.Labels["type"] == "pod_container"
				},
				Group: func(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set) {
					groupSet := batch.Sets[groupKey]
					if groupSet == nil {
						groupSet = &metrics.Set{
							Labels: map[string]string{},
							Values: map[string]metrics.Value{},
						}
					}
					return groupKey, groupSet
				},
			},
		})
		inputBatch := &metrics.Batch{
			Timestamp: time.Now(),
			Sets: map[metrics.ResourceKey]*metrics.Set{
				metrics.PodKey("ns1", "pod1"): {
					Labels: map[string]string{"type": "pod"},
					Values: map[string]metrics.Value{
						"m1": {
							ValueType: metrics.ValueInt64,
							IntValue:  1,
						},
						"c1": {
							ValueType: metrics.ValueInt64,
							IntValue:  2,
						},
					},
				},
				metrics.PodContainerKey("ns2", "pod2", "container2"): {
					Labels: map[string]string{"type": "pod_container"},
					Values: map[string]metrics.Value{
						"m2": {
							ValueType: metrics.ValueInt64,
							IntValue:  3,
						},
						"c2": {
							ValueType: metrics.ValueInt64,
							IntValue:  4,
						},
					},
				},
			},
		}

		outputBatch, err := sca.Process(inputBatch)
		assert.NoError(t, err)

		groupSet := outputBatch.Sets[groupKey]
		assert.NotNil(t, groupSet, "groupSet should exist")

		assert.Equal(t, int64(1), groupSet.Values["m1"].IntValue)
		assert.Equal(t, int64(2), groupSet.Values["c1"].IntValue)

		assert.Equal(t, int64(3), groupSet.Values["m2"].IntValue)
		assert.Equal(t, int64(4), groupSet.Values["c2"].IntValue)
	})
}
