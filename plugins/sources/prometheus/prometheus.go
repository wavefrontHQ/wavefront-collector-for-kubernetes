package prometheus

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/httputil"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/leadership"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	log "github.com/sirupsen/logrus"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	gometrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
)

var (
	collectErrors   gometrics.Counter
	filteredPoints  gometrics.Counter
	collectedPoints gometrics.Counter
)

func init() {
	pt := map[string]string{"type": "prometheus"}
	collectedPoints = gometrics.GetOrRegisterCounter(reporting.EncodeKey("source.points.collected", pt), gometrics.DefaultRegistry)
	filteredPoints = gometrics.GetOrRegisterCounter(reporting.EncodeKey("source.points.filtered", pt), gometrics.DefaultRegistry)
	collectErrors = gometrics.GetOrRegisterCounter(reporting.EncodeKey("source.collect.errors", pt), gometrics.DefaultRegistry)
}

type prometheusMetricsSource struct {
	metricsURL string
	prefix     string
	source     string
	tags       map[string]string
	buf        *bytes.Buffer
	filters    filter.Filter
	client     *http.Client
	pps        gometrics.Counter
	eps        gometrics.Counter
}

//TODO: move tags, prefix, source, filters into a single common struct used by all sources and sinks
func NewPrometheusMetricsSource(metricsURL, prefix, source, discovered string, tags map[string]string, filters filter.Filter, httpCfg httputil.ClientConfig) (metrics.MetricsSource, error) {
	client, err := httpClient(metricsURL, httpCfg)
	if err != nil {
		log.Errorf("error creating http client: %q", err)
		return nil, err
	}

	pt := extractTags(tags, discovered, metricsURL)
	ppsKey := reporting.EncodeKey("target.points.collected", pt)
	epsKey := reporting.EncodeKey("target.collect.errors", pt)

	return &prometheusMetricsSource{
		metricsURL: metricsURL,
		prefix:     prefix,
		source:     source,
		tags:       tags,
		buf:        bytes.NewBufferString(""),
		filters:    filters,
		client:     client,
		pps:        gometrics.GetOrRegisterCounter(ppsKey, gometrics.DefaultRegistry),
		eps:        gometrics.GetOrRegisterCounter(epsKey, gometrics.DefaultRegistry),
	}, nil
}

func extractTags(tags map[string]string, discovered, metricsURL string) map[string]string {
	result := make(map[string]string)
	for k, v := range tags {
		if k == "pod" || k == "service" || k == "apiserver" || k == "namespace" || k == "node" {
			result[k] = v
		}
	}
	if discovered != "" {
		result["discovered"] = discovered
	} else {
		result["discovered"] = "static"
		result["url"] = metricsURL
	}
	result["type"] = "prometheus"
	return result
}

func httpClient(metricsURL string, cfg httputil.ClientConfig) (*http.Client, error) {
	if strings.Contains(metricsURL, "kubernetes.default.svc.cluster.local") {
		if cfg.TLSConfig.CAFile == "" {
			log.Debugf("using default client for kubernetes api service")
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

func (src *prometheusMetricsSource) ScrapeMetrics() (*metrics.DataBatch, error) {
	result := &metrics.DataBatch{
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

func (src *prometheusMetricsSource) parseMetrics(buf []byte, header http.Header) ([]*metrics.MetricPoint, error) {
	var parser expfmt.TextParser

	// parse even if the buffer begins with a newline
	buf = bytes.TrimPrefix(buf, []byte("\n"))
	// Read raw data
	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)

	metricFamilies, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		log.Errorf("reading text format failed: %s", err)
	}
	return src.buildPoints(metricFamilies)
}

func (src *prometheusMetricsSource) buildPoints(metricFamilies map[string]*dto.MetricFamily) ([]*metrics.MetricPoint, error) {
	now := time.Now().Unix()
	var result []*metrics.MetricPoint

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
	log.Debugf("%s total points: %d", src.Name(), len(result))
	return result, nil
}

func (src *prometheusMetricsSource) metricPoint(name string, value float64, ts int64, source string, tags string) *metrics.MetricPoint {
	return &metrics.MetricPoint{
		Metric:    src.prefix + strings.Replace(name, "_", ".", -1),
		Value:     value,
		Timestamp: ts,
		Source:    source,
		StrTags:   tags,
	}
}

// Get name and value from metric
func (src *prometheusMetricsSource) buildPoint(name string, m *dto.Metric, now int64, tags string) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
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
func (src *prometheusMetricsSource) buildQuantiles(name string, m *dto.Metric, now int64, tags string) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
	for _, q := range m.GetSummary().Quantile {
		if !math.IsNaN(q.GetValue()) {
			newTags := fmt.Sprintf("%s quantile=%v", tags, q.GetQuantile())
			point := src.metricPoint(name, float64(q.GetValue()), now, src.source, newTags)
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
func (src *prometheusMetricsSource) buildHistos(name string, m *dto.Metric, now int64, tags string) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
	for _, b := range m.GetHistogram().Bucket {
		newTags := fmt.Sprintf("%s bucket=%v", tags, b.GetUpperBound())
		point := src.metricPoint(name, float64(b.GetCumulativeCount()), now, src.source, newTags)
		result = src.filterAppend(result, point)
	}
	point := src.metricPoint(name+".count", float64(m.GetHistogram().GetSampleCount()), now, src.source, tags)
	result = src.filterAppend(result, point)
	point = src.metricPoint(name+".sum", float64(m.GetHistogram().GetSampleSum()), now, src.source, tags)
	result = src.filterAppend(result, point)
	return result
}

// Get labels from metric
func (src *prometheusMetricsSource) buildTags(m *dto.Metric) string {
	src.buf.Reset()
	encodeTags(src.tags, src.buf)
	encodeLabelTags(m.Label, src.buf)
	return src.buf.String()
}

func encodeLabelTags(labels []*dto.LabelPair, buf *bytes.Buffer) {
	if len(labels) >= 0 {
		for _, label := range labels {
			if label.GetName() != "" && label.GetValue() != "" {
				buf.WriteString(" ")
				buf.WriteString(url.QueryEscape(label.GetName()))
				buf.WriteString("=")
				buf.WriteString(url.QueryEscape(label.GetValue()))
			}
		}
	}
}

func encodeTags(tags map[string]string, buf *bytes.Buffer) {
	for k, v := range tags {
		if len(v) > 0 {
			buf.WriteString(" ")
			buf.WriteString(url.QueryEscape(k))
			buf.WriteString("=")
			buf.WriteString(url.QueryEscape(v))
		}
	}
}

func (src *prometheusMetricsSource) filterAppend(slice []*metrics.MetricPoint, point *metrics.MetricPoint) []*metrics.MetricPoint {
	if src.filters == nil || src.filters.Match(point.Metric, point.Tags) {
		return append(slice, point)
	}
	filteredPoints.Inc(1)
	log.Debugf("dropping metric: %s", point.Metric)
	return slice
}

type prometheusProvider struct {
	metrics.DefaultMetricsSourceProvider
	urls       []string
	prefix     string
	source     string
	name       string
	discovered string
	sources    []metrics.MetricsSource
	tags       map[string]string
	filters    filter.Filter
}

func (p *prometheusProvider) GetMetricsSources() []metrics.MetricsSource {
	if p.discovered == "" && !leadership.Leading() {
		log.Infof("not scraping sources from: %s. current leader: %s", p.name, leadership.Leader())
		return nil
	}
	return p.sources
}

func (p *prometheusProvider) Name() string {
	return p.name
}

const providerName = "prometheus_metrics_provider"

func NewPrometheusProvider(cfg configuration.PrometheusSourceConfig) (metrics.MetricsSourceProvider, error) {
	if len(cfg.URL) == 0 {
		return nil, fmt.Errorf("missing prometheus url")
	}

	source := configuration.GetStringValue(cfg.Source, util.GetNodeName())
	source = configuration.GetStringValue(source, "prom_source")

	name := ""
	if len(cfg.Name) > 0 {
		name = fmt.Sprintf("%s: %s", providerName, cfg.Name)
	}
	if name == "" {
		name = fmt.Sprintf("%s: %s", providerName, cfg.URL)
	}

	discovered := configuration.GetStringValue(cfg.Discovered, "")
	log.Debugf("name: %s discovered: %s", name, discovered)

	httpCfg := cfg.HTTPClientConfig
	prefix := cfg.Prefix
	tags := cfg.Tags
	filters := filter.FromConfig(cfg.Filters)

	var sources []metrics.MetricsSource
	metricsSource, err := NewPrometheusMetricsSource(cfg.URL, prefix, source, discovered, tags, filters, httpCfg)
	if err == nil {
		sources = append(sources, metricsSource)
	} else {
		log.Errorf("error creating source: %v", err)
	}

	return &prometheusProvider{
		urls:       []string{cfg.URL},
		prefix:     prefix,
		source:     source,
		name:       name,
		discovered: discovered,
		sources:    sources,
		tags:       tags,
		filters:    filters,
	}, nil
}
