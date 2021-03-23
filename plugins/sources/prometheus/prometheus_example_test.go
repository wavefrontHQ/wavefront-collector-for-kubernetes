// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"sort"
	"testing"
)

// example pulled from https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md

func TestParsingOfCounterPoints(t *testing.T) {
	src := &prometheusMetricsSource{}

	points, err := src.parseMetrics(bytes.NewReader([]byte(`
# HELP http_requests_total The total number of HTTP requests.
# TYPE http_requests_total counter
http_requests_total{method="post",code="200"} 1027 1395066363000
http_requests_total{method="post",code="400"}    3 1395066363000
`)))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 2, len(points))
	sort.Sort(byKeyValue(points))

	assert.Equal(t, "http.requests.total.counter", points[0].Metric)
	assert.Equal(t, float64(3), points[0].Value)
	assert.Equal(t, map[string]string{"method": "post", "code": "400"}, points[0].GetTags(), "wrong point tags")

	assert.Equal(t, "http.requests.total.counter", points[1].Metric)
	assert.Equal(t, float64(1027), points[1].Value)
	assert.Equal(t, map[string]string{"method": "post", "code": "200"}, points[1].GetTags(), "wrong point tags")
}

func TestParsingOfHistogramPoints(t *testing.T) {
	src := &prometheusMetricsSource{}

	points, err := src.parseMetrics(bytes.NewReader([]byte(`
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
`)))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, 8, len(points))
	sort.Sort(byKeyValue(points))

	assert.Equal(t, "http.request.duration.seconds.bucket", points[0].Metric)
	assert.Equal(t, float64(24054), points[0].Value)
	assert.Equal(t, map[string]string{"le": "0.05"}, points[0].GetTags(), "wrong point tags")

	assert.Equal(t, "http.request.duration.seconds.bucket", points[5].Metric)
	assert.Equal(t, float64(144320), points[5].Value)
	assert.Equal(t, map[string]string{"le": "+Inf"}, points[5].GetTags(), "wrong point tags")

	assert.Equal(t, "http.request.duration.seconds.count", points[6].Metric)
	assert.Equal(t, float64(144320), points[6].Value)
	assert.Equal(t, make(map[string]string), points[6].GetTags(), "wrong point tags")

	assert.Equal(t, "http.request.duration.seconds.sum", points[7].Metric)
	assert.Equal(t, float64(53423), points[7].Value)
	assert.Equal(t, make(map[string]string), points[7].GetTags(), "wrong point tags")
}

func TestParsingOfQuantilePoints(t *testing.T) {
	src := &prometheusMetricsSource{}

	points, err := src.parseMetrics(bytes.NewReader([]byte(`
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
`)))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, 7, len(points))
	sort.Sort(byKeyValue(points))

	assert.Equal(t, "rpc.duration.seconds", points[0].Metric)
	assert.Equal(t, float64(3102), points[0].Value)
	assert.Equal(t, map[string]string{"quantile": "0.01"}, points[0].GetTags(), "wrong point tags")

	assert.Equal(t, "rpc.duration.seconds", points[4].Metric)
	assert.Equal(t, float64(76656), points[4].Value)
	assert.Equal(t, map[string]string{"quantile": "0.99"}, points[4].GetTags(), "wrong point tags")

	assert.Equal(t, "rpc.duration.seconds.count", points[5].Metric)
	assert.Equal(t, float64(2693), points[5].Value)
	assert.Equal(t, make(map[string]string), points[5].GetTags(), "wrong point tags")

	assert.Equal(t, "rpc.duration.seconds.sum", points[6].Metric)
	assert.Equal(t, 1.7560473e+07, points[6].Value)
	assert.Equal(t, make(map[string]string), points[6].GetTags(), "wrong point tags")
}

type byKeyValue []*metrics.MetricPoint

func (a byKeyValue) Len() int      { return len(a) }
func (a byKeyValue) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byKeyValue) Less(i, j int) bool {
	if a[i].Metric == a[j].Metric {
		return a[i].Value < a[j].Value
	} else {
		return a[i].Metric < a[j].Metric
	}
}
