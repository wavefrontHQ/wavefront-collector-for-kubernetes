// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
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
	replacer   *strings.Replacer
	client     *http.Client
	pps        gometrics.Counter
	eps        gometrics.Counter
}

func NewPrometheusMetricsSource(metricsURL, discovered string, transforms configuration.Transforms, httpCfg httputil.ClientConfig) (metrics.MetricsSource, error) {
	client, err := httpClient(metricsURL, httpCfg)
	if err != nil {
		log.Errorf("error creating http client: %q", err)
		return nil, err
	}

	pt := extractTags(transforms.Tags, discovered, metricsURL)
	ppsKey := reporting.EncodeKey("target.points.collected", pt)
	epsKey := reporting.EncodeKey("target.collect.errors", pt)

	replacer := strings.NewReplacer("_", ".")
	if transforms.ConvertPaths != nil && !*transforms.ConvertPaths {
		log.Debugf("retaining underscores for %s", metricsURL)
		replacer = strings.NewReplacer("_", "_")
	}

	return &prometheusMetricsSource{
		metricsURL: metricsURL,
		prefix:     transforms.Prefix,
		source:     transforms.Source,
		tags:       transforms.Tags,
		buf:        bytes.NewBufferString(""),
		filters:    filter.FromConfig(transforms.Filters),
		replacer:   replacer,
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
			var points []*metrics.MetricPoint
			if mf.GetType() == dto.MetricType_SUMMARY {
				// summary point
				points = src.buildQuantiles(metricName, m, now, tags)
			} else if mf.GetType() == dto.MetricType_HISTOGRAM {
				// histogram point
				points = src.buildHistos(metricName, m, now, tags)
			} else {
				// standard point
				points = src.buildPoint(metricName, m, now, tags)
			}

			for _, point := range points {
				if src.isValidMetric(point.Metric, point.Tags) {
					point.StrTags = src.encodeTags(point.Tags)
					point.Tags = nil
					result = append(result, point)
				}
			}
		}
	}
	log.Debugf("%s total points: %d", src.Name(), len(result))
	return result, nil
}

func (src *prometheusMetricsSource) metricPoint(name string, value float64, ts int64, source string, tags map[string]string) *metrics.MetricPoint {
	return &metrics.MetricPoint{
		Metric:    src.prefix + src.replacer.Replace(name),
		Value:     value,
		Timestamp: ts,
		Source:    source,
		Tags:      tags,
	}
}

// Get name and value from metric
func (src *prometheusMetricsSource) buildPoint(name string, m *dto.Metric, now int64, tags map[string]string) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			point := src.metricPoint(name+".gauge", float64(m.GetGauge().GetValue()), now, src.source, tags)
			result = append(result, point)
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			point := src.metricPoint(name+".counter", float64(m.GetCounter().GetValue()), now, src.source, tags)
			result = append(result, point)
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			point := src.metricPoint(name+".value", float64(m.GetUntyped().GetValue()), now, src.source, tags)
			result = append(result, point)
		}
	}
	return result
}

// Get Quantiles from summary metric
func (src *prometheusMetricsSource) buildQuantiles(name string, m *dto.Metric, now int64, tags map[string]string) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
	for _, q := range m.GetSummary().Quantile {
		if !math.IsNaN(q.GetValue()) {
			newTags := combineTags(tags, "quantile", fmt.Sprintf("%v", q.GetQuantile()))
			point := src.metricPoint(name, float64(q.GetValue()), now, src.source, newTags)
			result = append(result, point)
		}
	}
	point := src.metricPoint(name+".count", float64(m.GetSummary().GetSampleCount()), now, src.source, tags)
	result = append(result, point)
	point = src.metricPoint(name+".sum", float64(m.GetSummary().GetSampleSum()), now, src.source, tags)
	result = append(result, point)

	return result
}

// Get Buckets from histogram metric
func (src *prometheusMetricsSource) buildHistos(name string, m *dto.Metric, now int64, tags map[string]string) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
	for _, b := range m.GetHistogram().Bucket {
		newTags := combineTags(tags, "le", fmt.Sprintf("%v", b.GetUpperBound()))
		point := src.metricPoint(name, float64(b.GetCumulativeCount()), now, src.source, newTags)
		result = append(result, point)
	}
	point := src.metricPoint(name+".count", float64(m.GetHistogram().GetSampleCount()), now, src.source, tags)
	result = append(result, point)
	point = src.metricPoint(name+".sum", float64(m.GetHistogram().GetSampleSum()), now, src.source, tags)
	result = append(result, point)
	return result
}

// Get labels from metric
func (src *prometheusMetricsSource) buildTags(m *dto.Metric) map[string]string {
	tags := make(map[string]string, len(src.tags)+len(m.Label))
	for k, v := range src.tags {
		if len(v) > 0 {
			tags[k] = v
		}
	}
	if len(m.Label) >= 0 {
		for _, label := range m.Label {
			if len(label.GetName()) > 0 && len(label.GetValue()) > 0 {
				tags[label.GetName()] = label.GetValue()
			}
		}
	}
	return tags
}

func (src *prometheusMetricsSource) encodeTags(tags map[string]string) string {
	src.buf.Reset()
	for k, v := range tags {
		if len(v) > 0 {
			src.buf.WriteString(" ")
			src.buf.WriteString(url.QueryEscape(k))
			src.buf.WriteString("=")
			src.buf.WriteString(url.QueryEscape(v))
		}
	}
	return src.buf.String()
}

func (src *prometheusMetricsSource) isValidMetric(name string, tags map[string]string) bool {
	if src.filters == nil || src.filters.Match(name, tags) {
		return true
	}
	filteredPoints.Inc(1)
	log.Tracef("dropping metric: %s", name)
	return false
}

func combineTags(tags map[string]string, key, val string) map[string]string {
	newTags := make(map[string]string, len(tags)+1)
	for k, v := range tags {
		newTags[k] = v
	}
	newTags[key] = val
	return newTags
}

type prometheusProvider struct {
	metrics.DefaultMetricsSourceProvider
	name              string
	useLeaderElection bool
	sources           []metrics.MetricsSource
}

func (p *prometheusProvider) GetMetricsSources() []metrics.MetricsSource {
	if p.useLeaderElection && !leadership.Leading() {
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
	cfg.Transforms.Source = source

	if cfg.ConvertPaths == nil {
		convert, _ := strconv.ParseBool(os.Getenv("convertPaths"))
		cfg.Transforms.ConvertPaths = &convert
	}

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

	var sources []metrics.MetricsSource
	metricsSource, err := NewPrometheusMetricsSource(cfg.URL, discovered, cfg.Transforms, httpCfg)
	if err == nil {
		sources = append(sources, metricsSource)
	} else {
		return nil, fmt.Errorf("error creating source: %v", err)
	}

	useLeaderElection := cfg.UseLeaderElection || discovered == ""

	return &prometheusProvider{
		name:              name,
		useLeaderElection: useLeaderElection,
		sources:           sources,
	}, nil
}
