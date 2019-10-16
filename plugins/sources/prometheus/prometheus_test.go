package prometheus

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
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

func TestNoFilters(t *testing.T) {
	src := &prometheusMetricsSource{
		buf: bytes.NewBufferString(""),
	}

	metrics, err := src.parseMetrics([]byte(metricsStr), nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 8, len(metrics), "wrong number of metrics")
}

func TestMetricWhitelist(t *testing.T) {
	cfg := filter.Config{
		MetricWhitelist: []string{"*seconds.count*"},
	}
	f := filter.NewGlobFilter(cfg)

	src := &prometheusMetricsSource{
		buf:     bytes.NewBufferString(""),
		filters: f,
	}

	metrics, err := src.parseMetrics([]byte(metricsStr), nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 1, len(metrics), "wrong number of metrics")
}

func TestMetricBlacklist(t *testing.T) {
	cfg := filter.Config{
		MetricBlacklist: []string{"*seconds.count*"},
	}
	f := filter.NewGlobFilter(cfg)

	src := &prometheusMetricsSource{
		buf:     bytes.NewBufferString(""),
		filters: f,
	}

	metrics, err := src.parseMetrics([]byte(metricsStr), nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 7, len(metrics), "wrong number of metrics")
}

func TestMetricTagWhitelist(t *testing.T) {
	cfg := filter.Config{
		MetricTagWhitelist: map[string][]string{"label": {"good"}},
	}
	f := filter.NewGlobFilter(cfg)

	src := &prometheusMetricsSource{
		buf:     bytes.NewBufferString(""),
		filters: f,
	}

	metrics, err := src.parseMetrics([]byte(metricsStr), nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 1, len(metrics), "wrong number of metrics")
}

func TestMetricTagBlacklist(t *testing.T) {
	cfg := filter.Config{
		MetricTagBlacklist: map[string][]string{"label": {"ba*"}},
	}
	f := filter.NewGlobFilter(cfg)

	src := &prometheusMetricsSource{
		buf:     bytes.NewBufferString(""),
		filters: f,
	}

	metrics, err := src.parseMetrics([]byte(metricsStr), nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 7, len(metrics), "wrong number of metrics")
}
