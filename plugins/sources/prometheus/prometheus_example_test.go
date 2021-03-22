package prometheus

import (
	"bytes"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// example pulled from https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md
var _ = Describe("Prometheus Metric Examples", func() {
	It("parses simple counters", func() {
		metricsStr := `
# HELP http_requests_total The total number of HTTP requests.
# TYPE http_requests_total counter
http_requests_total{method="post",code="200"} 1027 1395066363000
http_requests_total{method="post",code="400"}    3 1395066363000`

		src := &prometheusMetricsSource{}
		points, err := src.parseMetrics(bytes.NewReader([]byte(metricsStr)))
		if err != nil {
			Fail(err.Error())
		}
		Expect(points).Should(HaveLen(2))
		sort.Sort(byValue(points))

		Expect(points[0].Metric).Should(Equal("http.requests.total.counter"))
		Expect(points[0].Value).Should(Equal(float64(1027)))
		Expect(points[0].GetTags()).Should(Equal(map[string]string{"method": "post", "code": "200"}))

		Expect(points[1].Metric).Should(Equal("http.requests.total.counter"))
		Expect(points[1].Value).Should(Equal(float64(3)))
		Expect(points[1].GetTags()).Should(Equal(map[string]string{"method": "post", "code": "400"}))
	})

	It("parses histograms", func() {
		metricsStr := `
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
`
		src := &prometheusMetricsSource{ /*tags: map[string]string {"pod":"myPod"}*/ }
		points, err := src.parseMetrics(bytes.NewReader([]byte(metricsStr)))
		if err != nil {
			Fail(err.Error())
		}
		Expect(points).Should(HaveLen(8))
		Expect(points).Should(Equal(nil))
		Expect(points[0].Metric).Should(Equal("http.requests.duration.seconds.bucket"))
		Expect(points[7].Metric).Should(Equal("http.requests.duration.seconds.sum"))
		sort.Sort(byValue(points))
	})
})

//func testExamplePromMetricsReader() *bytes.Reader {
const metricsStr = `
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
`

type byValue []*metrics.MetricPoint

func (a byValue) Len() int           { return len(a) }
func (a byValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byValue) Less(i, j int) bool { return a[i].Value > a[j].Value }
