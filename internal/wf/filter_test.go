package wf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
)

func TestFilterAppend(t *testing.T) {
	t.Run("does not add the point when Filter returns nil", func(t *testing.T) {
		expectedPoint := NewPoint("some.metric", 1.0, 2, "pod-123", map[string]string{"foo": "bar"})
		incrementer := fakeCounter(0)
		filters := filter.NewGlobFilter(filter.Config{
			MetricDenyList: []string{"some*"},
		})

		actualPoints := FilterAppend(filters, &incrementer, []Metric{}, expectedPoint)

		assert.Equal(t, fakeCounter(1), incrementer, "increments filtered")
		assert.Equal(t, 0, len(actualPoints), "does not add the point")
	})

	t.Run("adds the point when Filter returns a point", func(t *testing.T) {
		expectedPoint := NewPoint("some.metric", 1.0, 2, "pod-123", map[string]string{"foo": "bar"})
		incrementer := fakeCounter(0)

		actualPoints := FilterAppend(nil, &incrementer, []Metric{}, expectedPoint)

		assert.Equal(t, fakeCounter(0), incrementer, "does not increment filtered")
		if assert.Equal(t, 1, len(actualPoints), "adds the point") {
			assert.Equal(t, expectedPoint, actualPoints[0])
		}
	})
}

func TestFilter(t *testing.T) {
	t.Run("nil filter returns the point unmodified", func(t *testing.T) {
		expectedPoint := NewPoint("some.metric", 1.0, 2, "pod-123", map[string]string{"foo": "bar"})
		incrementer := fakeCounter(0)

		actualPoint := Filter(nil, &incrementer, expectedPoint)

		assert.Equal(t, fakeCounter(0), incrementer, "does not increment filtered")
		assert.Equal(t, expectedPoint, actualPoint, "returns the point unmodified")
	})

	t.Run("filters point", func(t *testing.T) {
		expectedPoint := NewPoint("some.metric", 1.0, 2, "pod-123", map[string]string{"foo": "bar"})
		incrementer := fakeCounter(0)
		filters := filter.NewGlobFilter(filter.Config{
			MetricDenyList: []string{"some*"},
		})

		actualPoint := Filter(filters, &incrementer, expectedPoint)

		assert.Equal(t, fakeCounter(1), incrementer, "increments filtered")
		assert.Nil(t, actualPoint, "returns nil point")
	})

	t.Run("filters tags on matched point", func(t *testing.T) {
		expectedPoint := NewPoint("some.metric", 1.0, 2, "pod-123", map[string]string{"foo": "bar", "bar": "foo"})
		incrementer := fakeCounter(0)
		filters := filter.NewGlobFilter(filter.Config{
			MetricDenyList: []string{"other*"},
			TagExclude:     []string{"foo*"},
		})

		actualPoint := Filter(filters, &incrementer, expectedPoint)

		assert.Equal(t, fakeCounter(0), incrementer, "does not increment filtered")
		if assert.Equal(t, actualPoint, expectedPoint, "returns the point") &&
			assert.Equal(t, 1, len(actualPoint.Tags()), "filters correct tags") {
			assert.Equal(t, "foo", actualPoint.Tags()["bar"], "preserves bar tag")
		}
	})

	t.Run("Does not try to filter on nil point", func(t *testing.T) {
		incrementer := fakeCounter(0)
		filters := filter.NewGlobFilter(filter.Config{
			MetricDenyList: []string{"other*"},
			TagExclude:     []string{"foo*"},
		})

		actualPoint := Filter(filters, &incrementer, nil)

		assert.Equal(t, fakeCounter(0), incrementer, "does not increment filtered")
		assert.Nil(t, actualPoint)
	})
}

type fakeCounter int64

func (f *fakeCounter) Inc(by int64) {
	*f += fakeCounter(by)
}
