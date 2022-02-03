package wf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
)

func TestFilterAppend(t *testing.T) {
	t.Run("nil filter acts like append", func(t *testing.T) {
		somePoint := NewPoint("some.metric", 1.0, 2, "pod-123", map[string]string{"foo": "bar"})
		incrementer := fakeCounter(0)

		points := FilterAppend(nil, &incrementer, []*Point{}, somePoint)

		assert.Equal(t, fakeCounter(0), incrementer, "does not increment filtered")
		if assert.Equal(t, 1, len(points), "adds the point to the list") {
			assert.Equal(t, somePoint, points[0], "adds the correct point to the list")
		}
	})

	t.Run("filters metrics", func(t *testing.T) {
		somePoint := NewPoint("some.metric", 1.0, 2, "pod-123", map[string]string{"foo": "bar"})
		incrementer := fakeCounter(0)
		filters := filter.NewGlobFilter(filter.Config{
			MetricDenyList: []string{"some*"},
		})

		points := FilterAppend(filters, &incrementer, []*Point{}, somePoint)

		assert.Equal(t, fakeCounter(1), incrementer, "increments filtered")
		assert.Equal(t, 0, len(points), "does not add the point to the list")
	})

	t.Run("filters tags on appended metrics", func(t *testing.T) {
		somePoint := NewPoint("some.metric", 1.0, 2, "pod-123", map[string]string{"foo": "bar", "bar": "foo"})
		incrementer := fakeCounter(0)
		filters := filter.NewGlobFilter(filter.Config{
			MetricDenyList: []string{"other*"},
			TagExclude:     []string{"foo*"},
		})

		points := FilterAppend(filters, &incrementer, []*Point{}, somePoint)

		assert.Equal(t, fakeCounter(0), incrementer, "does not increment filtered")
		if assert.Equal(t, 1, len(points), "adds the point to the list") &&
			assert.Equal(t, 1, len(points[0].Tags()), "filters correct tags") {
			assert.Equal(t, "foo", points[0].Tags()["bar"], "preserves bar tag")
		}
	})
}

type fakeCounter int64

func (f *fakeCounter) Inc(by int64) {
	*f += fakeCounter(by)
}
