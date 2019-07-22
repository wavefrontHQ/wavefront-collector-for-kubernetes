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
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/httputil"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	"github.com/golang/glog"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
)

var (
	collectErrors   metrics.Counter
	filteredPoints  metrics.Counter
	collectedPoints metrics.Counter
)

func init() {
	pt := map[string]string{"type": "prometheus"}
	collectedPoints = metrics.GetOrRegisterCounter(reporting.EncodeKey("source.points.collected", pt), metrics.DefaultRegistry)
	filteredPoints = metrics.GetOrRegisterCounter(reporting.EncodeKey("source.points.filtered", pt), metrics.DefaultRegistry)
	collectErrors = metrics.GetOrRegisterCounter(reporting.EncodeKey("source.collect.errors", pt), metrics.DefaultRegistry)
}

type prometheusMetricsSource struct {
	metricsURL string
	prefix     string
	source     string
	tags       map[string]string
	flatTags   string
	buf        *bytes.Buffer
	filters    filter.Filter
	client     *http.Client
	pps        metrics.Counter
	eps        metrics.Counter
}

//TODO: move tags, prefix, source, filters into a single common struct used by all sources and sinks
func NewPrometheusMetricsSource(metricsURL, prefix, source, discovered string, tags map[string]string, filters filter.Filter, httpCfg httputil.ClientConfig) (MetricsSource, error) {
	client, err := httpClient(metricsURL, httpCfg)
	if err != nil {
		glog.Errorf("error creating http client: %q", err)
		return nil, err
	}

	pt := extractTags(tags, discovered, metricsURL)
	ppsKey := reporting.EncodeKey("target.points.collected", pt)
	epsKey := reporting.EncodeKey("target.collect.errors", pt)
	flatTags := reporting.EncodeKey("", tags)

	return &prometheusMetricsSource{
		metricsURL: metricsURL,
		prefix:     prefix,
		source:     source,
		tags:       tags,
		flatTags:   flatTags,
		buf:        bytes.NewBufferString(""),
		filters:    filters,
		client:     client,
		pps:        metrics.GetOrRegisterCounter(ppsKey, metrics.DefaultRegistry),
		eps:        metrics.GetOrRegisterCounter(epsKey, metrics.DefaultRegistry),
	}, nil
}

func extractTags(tags map[string]string, discovered, metricsURL string) map[string]string {
	result := make(map[string]string)
	for k, v := range tags {
		if k == "pod" || k == "service" || k == "apiserver" || k == "namespace" {
			result[k] = v
		}
	}
	if discovered != "" {
		result["discovered"] = discovered
	} else {
		result["url"] = metricsURL
	}
	result["type"] = "prometheus"
	return result
}

func httpClient(metricsURL string, cfg httputil.ClientConfig) (*http.Client, error) {
	if strings.Contains(metricsURL, "kubernetes.default.svc.cluster.local") {
		if cfg.TLSConfig.CAFile == "" {
			glog.V(2).Info("using default client for kubernetes api service")
			cfg.TLSConfig.CAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
			cfg.TLSConfig.InsecureSkipVerify = true
		}
		if cfg.BearerToken == "" && cfg.BearerTokenFile == "" {
			cfg.BearerTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		}
	}
	client, err := httputil.NewClient(cfg)
	if err == nil {
		client.Timeout = time.Second * 30
	}
	return client, err
}

func (src *prometheusMetricsSource) Name() string {
	return fmt.Sprintf("prometheus_source: %s", src.metricsURL)
}

func (src *prometheusMetricsSource) ScrapeMetrics() (*DataBatch, error) {
	result := &DataBatch{
		Timestamp: time.Now(),
	}

	resp, err := src.client.Get(src.metricsURL)
	if err != nil {
		collectErrors.Inc(1)
		src.eps.Inc(1)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		collectErrors.Inc(1)
		src.eps.Inc(1)
		return nil, fmt.Errorf("error retrieving prometheus metrics from %s", src.metricsURL)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		collectErrors.Inc(1)
		src.eps.Inc(1)
		return nil, err
	}
	points, err := src.parseMetrics(body, resp.Header)
	if err != nil {
		collectErrors.Inc(1)
		src.eps.Inc(1)
		return result, err
	}
	result.MetricPoints = points
	collectedPoints.Inc(int64(len(points)))
	src.pps.Inc(int64(len(points)))

	return result, nil
}

func (src *prometheusMetricsSource) parseMetrics(buf []byte, header http.Header) ([]*MetricPoint, error) {
	var parser expfmt.TextParser

	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// Read raw data
	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)

	metricFamilies, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		glog.Errorf("reading text format failed: %s", err)
	}
	return src.buildPoints(metricFamilies)
}

func (src *prometheusMetricsSource) buildPoints(metricFamilies map[string]*dto.MetricFamily) ([]*MetricPoint, error) {
	now := time.Now().Unix()
	var result []*MetricPoint

	groupedTags := util.NewGroupedTags()

	for metricName, mf := range metricFamilies {
		for _, m := range mf.Metric {
			tags := src.buildTags(m, groupedTags)
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
func (src *prometheusMetricsSource) buildTags(m *dto.Metric, gt *util.GroupedTags) map[string]string {
	// always reset the buffer
	src.buf.Reset()

	// form a composite key of tags and labels of the form [key1=val1&key2=val2][label1=val1 label2=val2]
	src.buf.WriteString(src.flatTags)
	encodeLabelTags(m.Label, src.buf)
	key := src.buf.String()

	// check if a map of tags already exists for the key
	result, exists := gt.GetOrAdd(key)
	if exists {
		return result
	}

	// add to the map if no matching tags exist. The map reference is held within GroupedTags.
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

func encodeLabelTags(labels []*dto.LabelPair, buf *bytes.Buffer) {
	if len(labels) == 0 {
		return
	}
	buf.WriteString("[")
	for _, label := range labels {
		buf.WriteString(url.QueryEscape(label.GetName()))
		buf.WriteString("=")
		buf.WriteString(url.QueryEscape(label.GetValue()))
		buf.WriteString(url.QueryEscape(" "))
	}
	buf.WriteString("]")
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
	DefaultMetricsSourceProvider
	urls       []string
	prefix     string
	source     string
	name       string
	discovered string
	sources    []MetricsSource
	tags       map[string]string
	filters    filter.Filter
}

func (p *prometheusProvider) GetMetricsSources() []MetricsSource {
	if p.discovered == "" && !leadership.Leading() {
		glog.V(2).Infof("not scraping sources from: %s. current leader: %s", p.name, leadership.Leader())
		return nil
	}
	return p.sources
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

	source := flags.DecodeDefaultValue(vals, "source", util.GetNodeName())
	if source == "" {
		source = "prom_source"
	}

	name := ""
	if len(vals["name"]) > 0 {
		name = fmt.Sprintf("%s: %s", ProviderName, vals["name"][0])
	}
	if name == "" {
		name = fmt.Sprintf("%s: %s", ProviderName, vals["url"][0])
	}

	discovered := flags.DecodeValue(vals, "discovered")
	glog.V(4).Infof("name: %s discovered: %s", name, discovered)

	httpCfg := flags.DecodeHTTPConfig(vals)

	prefix := flags.DecodeValue(vals, "prefix")
	tags := flags.DecodeTags(vals)
	filters := filter.FromQuery(vals)

	var sources []MetricsSource
	for _, metricsURL := range vals["url"] {
		metricsSource, err := NewPrometheusMetricsSource(metricsURL, prefix, source, discovered, tags, filters, httpCfg)
		if err == nil {
			sources = append(sources, metricsSource)
		} else {
			glog.Errorf("error creating source: %v", err)
		}
	}

	return &prometheusProvider{
		urls:       vals["url"],
		prefix:     prefix,
		source:     source,
		name:       name,
		discovered: discovered,
		sources:    sources,
		tags:       tags,
		filters:    filters,
	}, nil
}
