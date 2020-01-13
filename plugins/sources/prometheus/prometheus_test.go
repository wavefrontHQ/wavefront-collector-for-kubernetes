package prometheus

import (
	"bytes"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/httputil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
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
		buf:      bytes.NewBufferString(""),
		replacer: strings.NewReplacer("_", "."),
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
		buf:      bytes.NewBufferString(""),
		filters:  f,
		replacer: strings.NewReplacer("_", "."),
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
		buf:      bytes.NewBufferString(""),
		filters:  f,
		replacer: strings.NewReplacer("_", "."),
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
		buf:      bytes.NewBufferString(""),
		filters:  f,
		replacer: strings.NewReplacer("_", "."),
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
		buf:      bytes.NewBufferString(""),
		filters:  f,
		replacer: strings.NewReplacer("_", "."),
	}

	metrics, err := src.parseMetrics([]byte(metricsStr), nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 7, len(metrics), "wrong number of metrics")
}

func TestConvertPaths(t *testing.T) {
	convert := true
	testConvertPaths(t, &convert)

	convert = false
	testConvertPaths(t, &convert)
}

func testConvertPaths(t *testing.T, convert *bool) {
	src, err := NewPrometheusMetricsSource(
		"http://testURL:8080",
		"",
		configuration.Transforms{
			ConvertPaths: convert,
		},
		httputil.ClientConfig{},
	)
	if err != nil {
		t.Errorf("error creating source")
	}
	ps := src.(*prometheusMetricsSource)

	points, err := ps.parseMetrics([]byte(metricsStr), nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 8, len(points), "wrong number of metrics")

	if !*convert {
		assert.True(t, strings.Contains(points[0].Metric, "http_request_duration_seconds"))
	} else {
		assert.True(t, strings.Contains(points[0].Metric, "http.request.duration.seconds"))
	}
}

func TestTransforms(t *testing.T) {
	convert := true
	p, err := NewPrometheusProvider(configuration.PrometheusSourceConfig{
		URL: "http://testURL:8080/metrics",
		Transforms: configuration.Transforms{
			Source: "testSource",
			Prefix: "testPrefix",
			Tags:   map[string]string{"env": "test", "from": "testTransforms"},
			Filters: filter.Config{
				MetricTagBlacklist: map[string][]string{"label": {"ba*"}},
			},
			ConvertPaths: &convert,
		},
	})
	if err != nil {
		t.Errorf("error creating prometheus provider: %v", err)
	}

	pp := p.(*prometheusProvider)
	assert.Equal(t, 1, len(pp.sources))

	src := pp.sources[0]
	ps := src.(*prometheusMetricsSource)
	assert.Equal(t, "testSource", ps.source)
	assert.Equal(t, "testPrefix", ps.prefix)
	assert.Equal(t, 2, len(ps.tags))
	assert.NotNil(t, ps.filters)
	assert.Equal(t, "test.metric.name", ps.replacer.Replace("test_metric_name"))
}
