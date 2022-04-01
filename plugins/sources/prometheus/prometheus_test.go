// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"bytes"
	"testing"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gm "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
)

func TestNoFilters(t *testing.T) {
	src := &prometheusMetricsSource{}
	batch := &metrics.Batch{
		Timestamp: time.Now(),
	}

	points, err := src.parseMetrics(testMetricReader(), batch)
	require.NoError(t, err, "parsing metrics")
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
	batch := &metrics.Batch{
		Timestamp: time.Now(),
	}

	points, err := src.parseMetrics(testMetricReader(), batch)
	require.NoError(t, err, "parsing metrics")
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
	batch := &metrics.Batch{
		Timestamp: time.Now(),
	}

	points, err := src.parseMetrics(testMetricReader(), batch)
	require.NoError(t, err, "parsing metrics")
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
	batch := &metrics.Batch{
		Timestamp: time.Now(),
	}

	points, err := src.parseMetrics(testMetricReader(), batch)
	require.NoError(t, err, "parsing metrics")
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
	batch := &metrics.Batch{
		Timestamp: time.Now(),
	}

	points, err := src.parseMetrics(testMetricReader(), batch)
	require.NoError(t, err, "parsing metrics")
	assert.Equal(t, 7, len(points), "wrong number of points")
}

func TestTagInclude(t *testing.T) {
	src := &prometheusMetricsSource{
		filters: filter.FromConfig(filter.Config{
			TagInclude: []string{"label"},
		}),
	}
	batch := &metrics.Batch{
		Timestamp: time.Now(),
	}

	points, err := src.parseMetrics(testMetricReader(), batch)
	require.NoError(t, err, "parsing metrics")
	assert.Equal(t, 8, len(points), "wrong number of points")

	tagCounts := map[string]int{}
	for _, point := range points {
		tags := point.Tags()
		for tagName := range tags {
			tagCounts[tagName] += 1
		}
	}
	assert.Equal(t, 1, len(tagCounts), "the only tags left are 'label'")
	assert.Equal(t, 2, tagCounts["label"], "two metrics have a tag named 'label'")
}

func TestTagExclude(t *testing.T) {
	src := &prometheusMetricsSource{
		filters: filter.FromConfig(filter.Config{
			TagExclude: []string{"label"},
		}),
	}

	batch := &metrics.Batch{
		Timestamp: time.Now(),
	}

	points, err := src.parseMetrics(testMetricReader(), batch)
	require.NoError(t, err, "parsing metrics")
	assert.Equal(t, 8, len(points), "wrong number of points")

	for _, point := range points {
		_, ok := point.Tags()["label"]
		assert.False(t, ok, point.Tags())
	}
}

func BenchmarkMetricPoint(b *testing.B) {
	filtered := gm.GetOrRegisterCounter("filtered", gm.DefaultRegistry)
	tempTags := map[string]string{"pod_name": "prometheus_pod_xyz", "namespace_name": "default"}
	src := &prometheusMetricsSource{prefix: "prefix."}
	pointBuilder := NewPointBuilder(src, filtered)
	for i := 0; i < b.N; i++ {
		_ = pointBuilder.point("http.requests.total.count", 1.0, 0, "", tempTags)
	}
}

func testMetricReader() *bytes.Reader {
	metricsStr := `
http_request_duration_seconds_bucket{le="0.5"} 0
http_request_duration_seconds_bucket{le="1"} 1
http_request_duration_seconds_bucket{le="2"} 2
http_request_duration_seconds_bucket{le="3"} 3
http_request_duration_seconds_bucket{le="5"} 3
http_request_duration_seconds_bucket{le="+Inf"} 3
http_request_duration_seconds_sum{label="bad"} 6
http_request_duration_seconds_count{label="good"} 3
`
	return bytes.NewReader([]byte(metricsStr))
}

func TestDiscoveredPrometheusMetricSource(t *testing.T) {
	t.Run("static source", func(t *testing.T) {
		ms, err := NewPrometheusMetricsSource("", "", "", "", map[string]string{}, nil, httputil.ClientConfig{})

		assert.Nil(t, err)
		assert.False(t, ms.AutoDiscovered(), "prometheus auto-discovery")
	})

	t.Run("discovered source", func(t *testing.T) {
		ms, err := NewPrometheusMetricsSource("", "", "", "some-discovery-method", map[string]string{}, nil, httputil.ClientConfig{})

		assert.Nil(t, err)
		assert.True(t, ms.AutoDiscovered(), "prometheus auto-discovery")
	})
}
