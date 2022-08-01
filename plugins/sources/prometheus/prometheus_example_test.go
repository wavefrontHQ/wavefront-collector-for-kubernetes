// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"bytes"
	"math"
	"sort"
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/experimental"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// example pulled from https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md

func TestParsingOfCounterPoints(t *testing.T) {
	src := &prometheusMetricsSource{}

	points := parsePoints(t, src, `
# HELP http_requests_total The total number of HTTP requests.
# TYPE http_requests_total counter
http_requests_total{method="post",code="200"} 1027 1395066363000
http_requests_total{method="post",code="400"}    3 1395066363000
`)

	assert.Equal(t, 2, len(points))

	assert.Equal(t, "http.requests.total.counter", points[0].Metric)
	assert.Equal(t, float64(3), points[0].Value)
	assert.Equal(t, map[string]string{"method": "post", "code": "400"}, points[0].Tags(), "wrong point tags")

	assert.Equal(t, "http.requests.total.counter", points[1].Metric)
	assert.Equal(t, float64(1027), points[1].Value)
	assert.Equal(t, map[string]string{"method": "post", "code": "200"}, points[1].Tags(), "wrong point tags")
}

func TestParsingOfHistogramPoints(t *testing.T) {
	src := &prometheusMetricsSource{}

	points := parsePoints(t, src, `
# A histogram, which has a pretty complex representation in the text format:
# HELP http_request_duration_seconds A histogram of the request duration.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.05"} 24054
http_request_duration_seconds_bucket{le="0.1"} 33444
http_request_duration_seconds_bucket{le="0.2"} 100392
http_request_duration_seconds_bucket{le="0.5"} 129389
http_request_duration_seconds_bucket{le="1"} 133988
http_request_duration_seconds_bucket{le="+Inf"} 144320
http_request_duration_seconds_sum 53423
http_request_duration_seconds_count 144320
`)

	assert.Equal(t, 8, len(points))

	assert.Equal(t, "http.request.duration.seconds.bucket", points[0].Metric)
	assert.Equal(t, float64(24054), points[0].Value)
	assert.Equal(t, map[string]string{"le": "0.05"}, points[0].Tags(), "wrong point tags")

	assert.Equal(t, "http.request.duration.seconds.bucket", points[5].Metric)
	assert.Equal(t, float64(144320), points[5].Value)
	assert.Equal(t, map[string]string{"le": "+Inf"}, points[5].Tags(), "wrong point tags")

	assert.Equal(t, "http.request.duration.seconds.count", points[6].Metric)
	assert.Equal(t, float64(144320), points[6].Value)

	assert.Equal(t, "http.request.duration.seconds.sum", points[7].Metric)
	assert.Equal(t, float64(53423), points[7].Value)
}

func TestParsingWFHistogramPoints(t *testing.T) {
	experimental.EnableFeature(experimental.HistogramConversion)
	src := &prometheusMetricsSource{}

	distributions := parseDistributions(t, src, `
# A histogram, which has a pretty complex representation in the text format:
# HELP http_request_duration_seconds A histogram of the request duration.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.05"} 24054
http_request_duration_seconds_bucket{le="0.2"} 100392
http_request_duration_seconds_bucket{le="0.1"} 33444
http_request_duration_seconds_bucket{le="0.5"} 129389
http_request_duration_seconds_bucket{le="1"} 133988
http_request_duration_seconds_bucket{le="+Inf"} 144320
http_request_duration_seconds_sum 53423
http_request_duration_seconds_count 144320
`)

	assert.Equal(t, 1, len(distributions))

	assert.Equal(t, 6, len(distributions[0].Centroids))
	expectedCentroids := []wf.Centroid{
		{Value: 0.05, Count: 24054},
		{Value: 0.1, Count: 33444},
		{Value: 0.2, Count: 100392},
		{Value: 0.5, Count: 129389},
		{Value: 1.0, Count: 133988},
		{Value: math.Inf(1), Count: 144320},
	}
	assert.Equal(t, expectedCentroids, distributions[0].Centroids)

	experimental.DisableFeature(experimental.HistogramConversion)
}

func TestParsingWFHistogramPointsWithOnlyInfiniteLE(t *testing.T) {
	experimental.EnableFeature(experimental.HistogramConversion)
	src := &prometheusMetricsSource{source: "somesource"}

	distributions := parseDistributions(t, src, `
# A histogram, which has a pretty complex representation in the text format:
# HELP http_request_duration_seconds A histogram of the request duration.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{sometag="somevalue", le="+Inf"} 144320
http_request_duration_seconds_sum{sometag="somevalue"} 50
http_request_duration_seconds_count{sometag="somevalue"} 144320
`)

	assert.Equal(t, 0, len(distributions))

	experimental.DisableFeature(experimental.HistogramConversion)
}

func TestParsingOfQuantilePoints(t *testing.T) {
	src := &prometheusMetricsSource{}

	points := parsePoints(t, src, `
# Finally a summary, which has a complex representation, too:
# HELP rpc_duration_seconds A summary of the RPC duration in seconds.
# TYPE rpc_duration_seconds summary
rpc_duration_seconds{quantile="0.01"} 3102
rpc_duration_seconds{quantile="0.05"} 3272
rpc_duration_seconds{quantile="0.5"} 4773
rpc_duration_seconds{quantile="0.9"} 9001
rpc_duration_seconds{quantile="0.99"} 76656
rpc_duration_seconds_sum 1.7560473e+07
rpc_duration_seconds_count 2693
`)

	assert.Equal(t, 7, len(points))

	assert.Equal(t, "rpc.duration.seconds", points[0].Metric)
	assert.Equal(t, float64(3102), points[0].Value)
	assert.Equal(t, map[string]string{"quantile": "0.01"}, points[0].Tags(), "wrong point tags")

	assert.Equal(t, "rpc.duration.seconds", points[4].Metric)
	assert.Equal(t, float64(76656), points[4].Value)
	assert.Equal(t, map[string]string{"quantile": "0.99"}, points[4].Tags(), "wrong point tags")

	assert.Equal(t, "rpc.duration.seconds.count", points[5].Metric)
	assert.Equal(t, float64(2693), points[5].Value)

	assert.Equal(t, "rpc.duration.seconds.sum", points[6].Metric)
	assert.Equal(t, 1.7560473e+07, points[6].Value)
}

func TestSourceTags(t *testing.T) {
	src := &prometheusMetricsSource{tags: map[string]string{"pod": "myPod"}}

	points := parsePoints(t, src, `
# HELP http_requests_total The total number of HTTP requests.
# TYPE http_requests_total counter
http_requests_total{method="post",code="200"} 1027 1395066363000
http_requests_total{method="post",code="400"}    3 1395066363000
`)

	assert.Equal(t, 2, len(points))
	assert.Equal(t, map[string]string{"method": "post", "code": "400", "pod": "myPod"}, points[0].Tags(), "wrong point tags")

	points = parsePoints(t, src, `
# A histogram, which has a pretty complex representation in the text format:
# HELP http_request_duration_seconds A histogram of the request duration.
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.05"} 24054
http_request_duration_seconds_bucket{le="0.1"} 33444
http_request_duration_seconds_bucket{le="0.2"} 100392
http_request_duration_seconds_bucket{le="0.5"} 129389
http_request_duration_seconds_bucket{le="1"} 133988
http_request_duration_seconds_bucket{le="+Inf"} 144320
http_request_duration_seconds_sum 53423
http_request_duration_seconds_count 144320
`)

	assert.Equal(t, 8, len(points))
	assert.Equal(t, map[string]string{"le": "0.05", "pod": "myPod"}, points[0].Tags(), "wrong point tags")
	assert.Equal(t, map[string]string{"pod": "myPod"}, points[6].Tags(), "wrong point tags for sum")
	assert.Equal(t, map[string]string{"pod": "myPod"}, points[6].Tags(), "wrong point tags for count")

	points = parsePoints(t, src, `
# Finally a summary, which has a complex representation, too:
# HELP rpc_duration_seconds A summary of the RPC duration in seconds.
# TYPE rpc_duration_seconds summary
rpc_duration_seconds{quantile="0.01"} 3102
rpc_duration_seconds{quantile="0.05"} 3272
rpc_duration_seconds{quantile="0.5"} 4773
rpc_duration_seconds{quantile="0.9"} 9001
rpc_duration_seconds{quantile="0.99"} 76656
rpc_duration_seconds_sum 1.7560473e+07
rpc_duration_seconds_count 2693
`)

	assert.Equal(t, 7, len(points))
	assert.Equal(t, map[string]string{"quantile": "0.01", "pod": "myPod"}, points[0].Tags(), "wrong point tags")
	assert.Equal(t, map[string]string{"pod": "myPod"}, points[5].Tags(), "wrong point tags for sum")
	assert.Equal(t, map[string]string{"pod": "myPod"}, points[6].Tags(), "wrong point tags for count")
}

func parsePoints(t *testing.T, src *prometheusMetricsSource, metricsString string) []*wf.Point {
	metrics, err := src.parseMetrics(bytes.NewReader([]byte(metricsString)))
	require.NoError(t, err, "parsing metrics")
	var points []*wf.Point
	for _, result := range metrics {
		point, ok := result.(*wf.Point)
		if ok {
			points = append(points, point)
		}
	}
	sort.Slice(points, func(i, j int) bool {
		a := points[i]
		b := points[j]
		if a.Metric == b.Metric {
			return a.Value < b.Value
		} else {
			return a.Metric < b.Metric
		}
	})
	return points
}

func parseDistributions(t *testing.T, src *prometheusMetricsSource, metricsString string) []*wf.Distribution {
	metrics, err := src.parseMetrics(bytes.NewReader([]byte(metricsString)))
	require.NoError(t, err, "parsing metrics")
	var distributions []*wf.Distribution
	for _, result := range metrics {
		distribution, ok := result.(*wf.Distribution)
		if ok {
			distributions = append(distributions, distribution)
		}
	}
	return distributions
}
