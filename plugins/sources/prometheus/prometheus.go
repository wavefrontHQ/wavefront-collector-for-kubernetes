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

	"github.com/golang/glog"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

type prometheusMetricsSource struct {
	metricsURL string
	prefix     string
	client     *http.Client
}

func NewPrometheusMetricsSource(metricsURL, prefix string) MetricsSource {
	return &prometheusMetricsSource{
		metricsURL: metricsURL,
		prefix:     prefix,
		client:     &http.Client{Timeout: time.Second * 10},
	}
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
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error retrieving prometheus metrics from %s", src.metricsURL)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	points, err := src.parseMetrics(body, resp.Header)
	if err != nil {
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
	result := []*MetricPoint{}

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

	glog.V(5).Infof("%s total points: %d", src.Name(), len(result))
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
			point := src.metricPoint(name+".gauge", float64(m.GetGauge().GetValue()), now, "prom_source", tags)
			result = append(result, point)
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			point := src.metricPoint(name+".counter", float64(m.GetCounter().GetValue()), now, "prom_source", tags)
			result = append(result, point)
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			point := src.metricPoint(name+".value", float64(m.GetUntyped().GetValue()), now, "prom_source", tags)
			result = append(result, point)
		}
	}
	return result
}

// Get Quantiles from summary metric
func (src *prometheusMetricsSource) buildQuantiles(name string, m *dto.Metric, now int64, tags map[string]string) []*MetricPoint {

	var result []*MetricPoint
	for _, q := range m.GetSummary().Quantile {
		if !math.IsNaN(q.GetValue()) {
			point := src.metricPoint(name+"."+fmt.Sprint(q.GetQuantile()), float64(q.GetValue()), now, "prom_source", tags)
			result = append(result, point)
		}
	}
	point := src.metricPoint(name+".count", float64(m.GetSummary().GetSampleCount()), now, "prom_source", tags)
	result = append(result, point)
	point = src.metricPoint(name+".sum", float64(m.GetSummary().GetSampleSum()), now, "prom_source", tags)
	result = append(result, point)

	return result
}

// Get Buckets from histogram metric
func (src *prometheusMetricsSource) buildHistos(name string, m *dto.Metric, now int64, tags map[string]string) []*MetricPoint {
	var result []*MetricPoint
	for _, b := range m.GetHistogram().Bucket {
		point := src.metricPoint(name+"."+fmt.Sprint(b.GetUpperBound()), float64(b.GetCumulativeCount()), now, "prom_source", tags)
		result = append(result, point)
	}
	point := src.metricPoint(name+".count", float64(m.GetHistogram().GetSampleCount()), now, "prom_source", tags)
	result = append(result, point)
	point = src.metricPoint(name+".sum", float64(m.GetHistogram().GetSampleSum()), now, "prom_source", tags)
	result = append(result, point)
	return result
}

// Get labels from metric
func (src *prometheusMetricsSource) buildTags(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, lp := range m.Label {
		result[lp.GetName()] = lp.GetValue()
	}
	return result
}

type prometheusProvider struct {
	urls   []string
	prefix string
}

func (p *prometheusProvider) GetMetricsSources() []MetricsSource {
	sources := []MetricsSource{}
	for _, metricsURL := range p.urls {
		sources = append(sources, NewPrometheusMetricsSource(metricsURL, p.prefix))
	}
	return sources
}

func (p *prometheusProvider) Name() string {
	return "Prometheus Metrics Provider"
}

func NewPrometheusProvider(uri *url.URL) (MetricsSourceProvider, error) {
	vals := uri.Query()

	if len(vals["url"]) == 0 {
		return nil, fmt.Errorf("missing prometheus url")
	}

	prefix := ""
	if len(vals["prefix"]) > 0 {
		prefix = vals["prefix"][0]
	}

	return &prometheusProvider{
		urls:   vals["url"],
		prefix: prefix,
	}, nil
}
