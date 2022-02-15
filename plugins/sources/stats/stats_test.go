package stats

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
)

func TestBuildTags(t *testing.T) {
	srcTags := map[string]string{"a": "a", "b": "b"}
	internalMetricSource := createInternalMetricsSource(configuration.StatsSourceConfig{}, srcTags)

	t.Run("combines internalMetricSource tags and passed in tags", func(t *testing.T) {
		tags := map[string]string{"c": "c"}
		pointTags := internalMetricSource.buildTags(tags)
		assert.True(t, reflect.DeepEqual(map[string]string{"a": "a", "b": "b", "c": "c"}, pointTags), "should combine tags")
	})

	t.Run("returns new map if passed in tags are empty to resolve sink / scrape map concurrency bug", func(t *testing.T) {
		tags := map[string]string{}
		pointTags := internalMetricSource.buildTags(tags)
		assert.True(t, reflect.DeepEqual(srcTags, pointTags), "should combine tags")
		assert.False(t, reflect.ValueOf(internalMetricSource.tags).Pointer() == reflect.ValueOf(pointTags).Pointer())
	})
	t.Run("return tags if src tags are empty", func(t *testing.T) {
		internalMetricSource = createInternalMetricsSource(configuration.StatsSourceConfig{}, map[string]string{})
		tags := map[string]string{"c": "c"}
		pointTags := internalMetricSource.buildTags(tags)
		assert.True(t, reflect.DeepEqual(tags, pointTags), "should return tags")
	})

}

func createInternalMetricsSource(cfg configuration.StatsSourceConfig, tags map[string]string) internalMetricsSource {
	prefix := configuration.GetStringValue(cfg.Prefix, "kubernetes.")
	filters := filter.FromConfig(cfg.Filters)

	src, _ := newInternalMetricsSource(prefix, tags, filters)
	return *src.(*internalMetricsSource)
}
