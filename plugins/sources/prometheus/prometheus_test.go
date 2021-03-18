package prometheus

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

var metricsStr = `
http_request_duration_seconds_bucket{le="0.5"} 0
http_request_duration_seconds_bucket{le="1"} 1
http_request_duration_seconds_bucket{le="2"} 2
http_request_duration_seconds_bucket{le="3"} 3
http_request_duration_seconds_bucket{le="5"} 3
http_request_duration_seconds_bucket{le="+Inf"} 3
http_request_duration_seconds_sum{label="bad"} 6
http_request_duration_seconds_count{label="good"} 3
`

func testMetricReader() *bytes.Reader {
	return bytes.NewReader([]byte(metricsStr))
}

func TestNoFilters(t *testing.T) {
	src := &prometheusMetricsSource{}

	points, err := src.parseMetrics(testMetricReader())
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 8, len(points), "wrong number of points")
}

func TestMetricAllowList(t *testing.T) {
	cfg := filter.Config{
		MetricAllowList: []string{"*seconds.count*"},
	}
	f := filter.FromConfig(cfg)

	src := &prometheusMetricsSource{
		filters: f,
	}

	points, err := src.parseMetrics(testMetricReader())
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 1, len(points), "wrong number of points")
}

func TestMetricDenyList(t *testing.T) {
	cfg := filter.Config{
		MetricDenyList: []string{"*seconds.count*"},
	}
	f := filter.FromConfig(cfg)

	src := &prometheusMetricsSource{
		filters: f,
	}

	points, err := src.parseMetrics(testMetricReader())
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 7, len(points), "wrong number of points")
}

func TestMetricTagAllowList(t *testing.T) {
	cfg := filter.Config{
		MetricTagAllowList: map[string][]string{"label": {"good"}},
	}
	f := filter.FromConfig(cfg)

	src := &prometheusMetricsSource{
		filters: f,
	}

	points, err := src.parseMetrics(testMetricReader())
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 1, len(points), "wrong number of points")
}

func TestMetricTagDenyList(t *testing.T) {
	cfg := filter.Config{
		MetricTagDenyList: map[string][]string{"label": {"ba*"}},
	}
	f := filter.FromConfig(cfg)

	src := &prometheusMetricsSource{
		filters: f,
	}

	points, err := src.parseMetrics(testMetricReader())
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 7, len(points), "wrong number of points")
}

func TestPointTags(t *testing.T) {
	src := &prometheusMetricsSource{}

	points, err := src.parseMetrics(bytes.NewReader([]byte(`http_request_duration_seconds_count{label="good"} 3`)))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, map[string]string{"label": "good"}, points[0].GetTags(), "wrong point tags")
}

var tempTags = map[string]string{"pod_name": "prometheus_pod_xyz", "namespace_name": "default"}
var result *metrics.MetricPoint

func BenchmarkMetricPoint(b *testing.B) {
	src := &prometheusMetricsSource{prefix: "prefix."}
	pointBuilder := NewPointBuilder(src)
	var r *metrics.MetricPoint
	for i := 0; i < b.N; i++ {
		r = pointBuilder.metricPoint("http.requests.total.count", 1.0, 0, "", tempTags)
	}
	result = r
}
