// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

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
	metricsURL           string
	prefix               string
	source               string
	tags                 map[string]string
	filters              filter.Filter
	client               *http.Client
	pps                  gometrics.Counter
	eps                  gometrics.Counter
	internalMetricsNames []string

	omitBucketSuffix bool
}

func NewPrometheusMetricsSource(metricsURL, prefix, source, discovered string, tags map[string]string, filters filter.Filter, httpCfg httputil.ClientConfig) (metrics.MetricsSource, error) {
	client, err := httpClient(metricsURL, httpCfg)
	if err != nil {
		log.Errorf("error creating http client: %q", err)
		return nil, err
	}

	pt := extractTags(tags, discovered, metricsURL)
	ppsKey := reporting.EncodeKey("target.points.collected", pt)
	epsKey := reporting.EncodeKey("target.collect.errors", pt)

	omitBucketSuffix, _ := strconv.ParseBool(os.Getenv("omitBucketSuffix"))

	return &prometheusMetricsSource{
		metricsURL:           metricsURL,
		prefix:               prefix,
		source:               source,
		tags:                 tags,
		filters:              filters,
		client:               client,
		pps:                  gometrics.GetOrRegisterCounter(ppsKey, gometrics.DefaultRegistry),
		eps:                  gometrics.GetOrRegisterCounter(epsKey, gometrics.DefaultRegistry),
		internalMetricsNames: []string{ppsKey, epsKey},
		omitBucketSuffix:     omitBucketSuffix,
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

func (src *prometheusMetricsSource) Cleanup() {
	for _, name := range src.internalMetricsNames {
		gometrics.Unregister(name)
	}
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
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		collectErrors.Inc(1)
		src.eps.Inc(1)
		return nil, fmt.Errorf("error retrieving prometheus metrics from %s", src.metricsURL)
	}

	points, err := src.parseMetrics(resp.Body)

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

func (src *prometheusMetricsSource) parseMetrics(reader io.Reader) ([]*metrics.MetricPoint, error) {

	metricReader := NewMetricReader(reader)

	var points = make([]*metrics.MetricPoint, 0)
	var err error
	for !metricReader.Done() {
		var parser expfmt.TextParser
		reader := bytes.NewReader(metricReader.Read())
		metricFamilies, err := parser.TextToMetricFamilies(reader)
		if err != nil {
			log.Errorf("reading text format failed: %s", err)
		}
		batch, err := src.buildPoints(metricFamilies)
		points = append(points, batch...)
	}
	return points, err
}

func (src *prometheusMetricsSource) buildPoints(metricFamilies map[string]*dto.MetricFamily) ([]*metrics.MetricPoint, error) {
	now := time.Now().Unix()
	var result []*metrics.MetricPoint

	for metricName, mf := range metricFamilies {
		for _, m := range mf.Metric {
			var points []*metrics.MetricPoint
			if mf.GetType() == dto.MetricType_SUMMARY {
				// summary point
				points = src.buildQuantiles(metricName, m, now, src.buildTags(m))
			} else if mf.GetType() == dto.MetricType_HISTOGRAM {
				// histogram point
				points = src.buildHistos(metricName, m, now, src.buildTags(m))
			} else {
				// standard point
				points = src.buildPoint(metricName, m, now)
			}

			if len(points) > 0 {
				result = append(result, points...)
			}
		}
	}
	return result, nil
}

func (src *prometheusMetricsSource) metricPoint(name string, value float64, ts int64, source string, tags map[string]string) *metrics.MetricPoint {
	return &metrics.MetricPoint{
		Metric:    src.prefix + strings.Replace(name, "_", ".", -1),
		Value:     value,
		Timestamp: ts,
		Source:    source,
		Tags:      tags,
	}
}

func (src *prometheusMetricsSource) filterAppend(slice []*metrics.MetricPoint, point *metrics.MetricPoint, m *dto.Metric) []*metrics.MetricPoint {
	// check whether we can avoid creating tags till after filtering.
	// basically if only metric name based filters are configured on the source
	// perform the filtering decision first and then create the tags
	if point.Tags == nil && (src.filters == nil || src.filters.UsesTags()) {
		point.Tags = src.buildTags(m)
	}

	if src.isValidMetric(point.Metric, point.Tags) {
		if point.Tags == nil {
			// skip allocating intermediate maps and pass label pairs and src tags directly to the sink
			point.Labels = m.Label
			point.SrcTags = src.tags
		}
		return append(slice, point)
	}
	return slice
}

// Get name and value from metric
func (src *prometheusMetricsSource) buildPoint(name string, m *dto.Metric, now int64) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
	if m.Gauge != nil {
		if !math.IsNaN(m.GetGauge().GetValue()) {
			point := src.metricPoint(name+".gauge", m.GetGauge().GetValue(), now, src.source, nil)
			result = src.filterAppend(result, point, m)
		}
	} else if m.Counter != nil {
		if !math.IsNaN(m.GetCounter().GetValue()) {
			point := src.metricPoint(name+".counter", m.GetCounter().GetValue(), now, src.source, nil)
			result = src.filterAppend(result, point, m)
		}
	} else if m.Untyped != nil {
		if !math.IsNaN(m.GetUntyped().GetValue()) {
			point := src.metricPoint(name+".value", m.GetUntyped().GetValue(), now, src.source, nil)
			result = src.filterAppend(result, point, m)
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
			point := src.metricPoint(name, q.GetValue(), now, src.source, newTags)
			result = src.filterAppend(result, point, m)
		}
	}
	point := src.metricPoint(name+".count", float64(m.GetSummary().GetSampleCount()), now, src.source, tags)
	result = src.filterAppend(result, point, m)
	point = src.metricPoint(name+".sum", m.GetSummary().GetSampleSum(), now, src.source, tags)
	result = src.filterAppend(result, point, m)

	return result
}

// Get Buckets from histogram metric
func (src *prometheusMetricsSource) buildHistos(name string, m *dto.Metric, now int64, tags map[string]string) []*metrics.MetricPoint {
	var result []*metrics.MetricPoint
	histName := src.histoName(name)
	for _, b := range m.GetHistogram().Bucket {
		newTags := combineTags(tags, "le", fmt.Sprintf("%v", b.GetUpperBound()))
		point := src.metricPoint(histName, float64(b.GetCumulativeCount()), now, src.source, newTags)
		result = src.filterAppend(result, point, m)
	}
	point := src.metricPoint(name+".count", float64(m.GetHistogram().GetSampleCount()), now, src.source, tags)
	result = src.filterAppend(result, point, m)
	point = src.metricPoint(name+".sum", m.GetHistogram().GetSampleSum(), now, src.source, tags)
	result = src.filterAppend(result, point, m)
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

func (src *prometheusMetricsSource) isValidMetric(name string, tags map[string]string) bool {
	if src.filters == nil || src.filters.Match(name, tags) {
		return true
	}
	filteredPoints.Inc(1)
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Tracef("dropping metric: %s", name)
	}
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

func (src *prometheusMetricsSource) histoName(name string) string {
	if src.omitBucketSuffix {
		return name
	}
	return name + ".bucket"
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
		return nil, fmt.Errorf("error creating source: %v", err)
	}

	useLeaderElection := cfg.UseLeaderElection || discovered == ""

	return &prometheusProvider{
		name:              name,
		useLeaderElection: useLeaderElection,
		sources:           sources,
	}, nil
}
