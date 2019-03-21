package prometheus

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/httputil"
	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"github.com/golang/glog"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/rcrowley/go-metrics"
)

var (
	scrapeErrors   metrics.Counter
	filteredPoints metrics.Counter
)

func init() {
	scrapeErrors = metrics.GetOrRegisterCounter("source.prometheus.scrape.errors", metrics.DefaultRegistry)
	filteredPoints = metrics.GetOrRegisterCounter("source.prometheus.points.filtered.count", metrics.DefaultRegistry)
}

type prometheusMetricsSource struct {
	metricsURL string
	prefix     string
	source     string
	tags       map[string]string
	filters    filter.Filter
	client     *http.Client
}

func NewPrometheusMetricsSource(metricsURL, prefix, source string, tags map[string]string, filters filter.Filter) (MetricsSource, error) {
	client, err := httpClient(metricsURL)
	if err != nil {
		glog.Errorf("error creating http client: %q", err)
		return nil, err
	}
	return &prometheusMetricsSource{
		metricsURL: metricsURL,
		prefix:     prefix,
		source:     source,
		tags:       tags,
		filters:    filters,
		client:     client,
	}, nil
}

func httpClient(metricsURL string) (*http.Client, error) {
	if strings.Contains(metricsURL, "kubernetes.default.svc.cluster.local") {
		client, err := httputil.NewClient(httputil.ClientConfig{
			TLSConfig: httputil.TLSConfig{
				CAFile:             "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
				InsecureSkipVerify: true,
			},
			BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		})
		glog.V(2).Info("using default client for kubernetes api service")
		return client, err
	}
	return &http.Client{Timeout: time.Second * 30}, nil
}

func (src *prometheusMetricsSource) Name() string {
	return fmt.Sprintf("prometheus_source: %s", src.metricsURL)
}

func (src *prometheusMetricsSource) ScrapeMetrics(start, end time.Time) (*DataBatch, error) {

	result := &DataBatch{
		Timestamp: time.Now(),
	}

	resp, err := src.client.Get(src.metricsURL)
	if err != nil {
		scrapeErrors.Inc(1)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		scrapeErrors.Inc(1)
		return nil, fmt.Errorf("error retrieving prometheus metrics from %s", src.metricsURL)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		scrapeErrors.Inc(1)
		return nil, err
	}
	points, err := src.parseMetrics(body, resp.Header)
	if err != nil {
		scrapeErrors.Inc(1)
		return result, err
	}
	result.MetricPoints = points

	return result, nil
}

func (src *prometheusMetricsSource) parseMetrics(buf []byte, header http.Header) ([]*MetricPoint, error) {
	var parser expfmt.TextParser

	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// Read raw data
	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)

	metricFamilies := make(map[string]*dto.MetricFamily)

	metricFamilies, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		glog.Errorf("reading text format failed: %s", err)
	}
	return src.buildPoints(metricFamilies)
}

func (src *prometheusMetricsSource) buildPoints(metricFamilies map[string]*dto.MetricFamily) ([]*MetricPoint, error) {
	now := time.Now().Unix()
	var result []*MetricPoint

	for metricName, mf := range metricFamilies {
		for _, m := range mf.Metric {
			tags := src.buildTags(m)
			if mf.GetType() == dto.MetricType_SUMMARY {
				// summary metric
				result = append(result, src.buildQuantiles(metricName, m, now, tags)...)
			} else if mf.GetType() == dto.MetricType_HISTOGRAM {
				// histogram metric
				result = append(result, src.buildHistos(metricName, m, now, tags)...)
			} else {
				// standard metric
				result = append(result, src.buildPoint(metricName, m, now, tags)...)
			}
		}
	}

	glog.V(4).Infof("%s total points: %d", src.Name(), len(result))
	if glog.V(9) {
		for _, i := range result {
			glog.Infof("%s %f src=%s %q \n", i.Metric, i.Value, i.Source, i.Tags)
		}
	}
	return result, nil
}

func (src *prometheusMetricsSource) metricPoint(name string, value float64, ts int64, source string, tags map[string]string) *MetricPoint {
	return &MetricPoint{
		Metric:    src.prefix + strings.Replace(name, "_", ".", -1),
		Value:     value,
		Timestamp: ts,
		Source:    source,
		Tags:      tags,
	}
}

// Get name and value from metric
func (src *prometheusMetricsSource) buildPoint(name string, m *dto.Metric, now int64, tags map[string]string) []*MetricPoint {
	var result []*MetricPoint
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			point := src.metricPoint(name+".gauge", float64(m.GetGauge().GetValue()), now, src.source, tags)
			result = src.filterAppend(result, point)
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			point := src.metricPoint(name+".counter", float64(m.GetCounter().GetValue()), now, src.source, tags)
			result = src.filterAppend(result, point)
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			point := src.metricPoint(name+".value", float64(m.GetUntyped().GetValue()), now, src.source, tags)
			result = src.filterAppend(result, point)
		}
	}
	return result
}

// Get Quantiles from summary metric
func (src *prometheusMetricsSource) buildQuantiles(name string, m *dto.Metric, now int64, tags map[string]string) []*MetricPoint {
	var result []*MetricPoint
	for _, q := range m.GetSummary().Quantile {
		if !math.IsNaN(q.GetValue()) {
			point := src.metricPoint(name+"."+fmt.Sprint(q.GetQuantile()), float64(q.GetValue()), now, src.source, tags)
			result = src.filterAppend(result, point)
		}
	}
	point := src.metricPoint(name+".count", float64(m.GetSummary().GetSampleCount()), now, src.source, tags)
	result = src.filterAppend(result, point)
	point = src.metricPoint(name+".sum", float64(m.GetSummary().GetSampleSum()), now, src.source, tags)
	result = src.filterAppend(result, point)

	return result
}

// Get Buckets from histogram metric
func (src *prometheusMetricsSource) buildHistos(name string, m *dto.Metric, now int64, tags map[string]string) []*MetricPoint {
	var result []*MetricPoint
	for _, b := range m.GetHistogram().Bucket {
		point := src.metricPoint(name+"."+fmt.Sprint(b.GetUpperBound()), float64(b.GetCumulativeCount()), now, src.source, tags)
		result = src.filterAppend(result, point)
	}
	point := src.metricPoint(name+".count", float64(m.GetHistogram().GetSampleCount()), now, src.source, tags)
	result = src.filterAppend(result, point)
	point = src.metricPoint(name+".sum", float64(m.GetHistogram().GetSampleSum()), now, src.source, tags)
	result = src.filterAppend(result, point)
	return result
}

// Get labels from metric
func (src *prometheusMetricsSource) buildTags(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for k, v := range src.tags {
		if len(v) > 0 {
			result[k] = v
		}
	}
	for _, lp := range m.Label {
		if len(lp.GetValue()) > 0 {
			result[lp.GetName()] = lp.GetValue()
		}
	}
	return result
}

func (src *prometheusMetricsSource) filterAppend(slice []*MetricPoint, point *MetricPoint) []*MetricPoint {
	if src.filters == nil || src.filters.Match(point.Metric, point.Tags) {
		return append(slice, point)
	}
	filteredPoints.Inc(1)
	glog.V(5).Infof("dropping metric: %s", point.Metric)
	return slice
}

type prometheusProvider struct {
	urls    []string
	prefix  string
	source  string
	name    string
	tags    map[string]string
	filters filter.Filter
}

func (p *prometheusProvider) GetMetricsSources() []MetricsSource {
	var sources []MetricsSource
	for _, metricsURL := range p.urls {
		source, err := NewPrometheusMetricsSource(metricsURL, p.prefix, p.source, p.tags, p.filters)
		if err == nil {
			sources = append(sources, source)
		}
	}
	return sources
}

func (p *prometheusProvider) Name() string {
	return p.name
}

const ProviderName = "prometheus_metrics_provider"

func NewPrometheusProvider(uri *url.URL) (MetricsSourceProvider, error) {
	vals := uri.Query()

	if len(vals["url"]) == 0 {
		return nil, fmt.Errorf("missing prometheus url")
	}

	prefix := ""
	if len(vals["prefix"]) > 0 {
		prefix = vals["prefix"][0]
	}

	source := "prom_source"
	if len(vals["source"]) > 0 {
		source = vals["source"][0]
	}

	name := ""
	if len(vals["name"]) > 0 {
		name = fmt.Sprintf("%s: %s", ProviderName, vals["name"][0])
	}
	if name == "" {
		name = fmt.Sprintf("%s: %s", ProviderName, vals["url"][0])
	}

	// tags of the form "tag=key:value"
	var tags map[string]string
	if len(vals["tag"]) > 0 {
		tags = make(map[string]string)
		tagList := vals["tag"]
		for _, tag := range tagList {
			s := strings.Split(tag, ":")
			if len(s) == 2 {
				k, v := s[0], s[1]
				tags[k] = v
			} else {
				glog.Warning("invalid tag ", tag)
			}
		}
	}

	filters := filter.FromQuery(vals)

	return &prometheusProvider{
		urls:    vals["url"],
		prefix:  prefix,
		source:  source,
		name:    name,
		tags:    tags,
		filters: filters,
	}, nil
}
