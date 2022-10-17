// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	log "github.com/sirupsen/logrus"

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
	autoDiscovered       bool

	omitBucketSuffix bool
}

func NewPrometheusMetricsSource(metricsURL, prefix, source, discovered string, tags map[string]string, filters filter.Filter, httpCfg httputil.ClientConfig) (metrics.Source, error) {
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
		autoDiscovered:       len(discovered) > 0,
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
	}
	client, err := httputil.NewClient(cfg)
	if err == nil {
		client.Timeout = time.Second * 30
	}
	return client, err
}

func (src *prometheusMetricsSource) AutoDiscovered() bool {
	return src.autoDiscovered
}

func (src *prometheusMetricsSource) Name() string {
	return fmt.Sprintf("prometheus_source: %s", src.metricsURL)
}

func (src *prometheusMetricsSource) Cleanup() {
	for _, name := range src.internalMetricsNames {
		gometrics.Unregister(name)
	}
}

type HTTPError struct {
	MetricsURL string
	Status     string
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("error retrieving prometheus metrics from %s (http status %s)", e.MetricsURL, e.Status)
}

func (src *prometheusMetricsSource) Scrape() (*metrics.Batch, error) {
	result := &metrics.Batch{
		Timestamp: time.Now(),
	}

	// TODO the likely reason this is not unit tested
	resp, err := src.client.Get(src.metricsURL)
	if err != nil {
		collectErrors.Inc(1)
		src.eps.Inc(1)
		return nil, err
	}

	/* TODO UNTESTED */
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		collectErrors.Inc(1)
		src.eps.Inc(1)
		return nil, &HTTPError{MetricsURL: src.metricsURL, Status: resp.Status, StatusCode: resp.StatusCode}
	}

	// TODO I can create a wrapper test class on prometheusMetricsSource and override parseMetrics to make testing easier!
	result.Metrics, err = src.parseMetrics(resp.Body)
	if err != nil {
		collectErrors.Inc(1)
		src.eps.Inc(1)
		return result, err
	}
	collectedPoints.Inc(int64(result.Points()))
	src.pps.Inc(int64(result.Points()))

	return result, nil
}

// parseMetrics converts serialized prometheus metrics to wavefront points
// parseMetrics returns an error when IO or parsing fails
func (src *prometheusMetricsSource) parseMetrics(reader io.Reader) ([]wf.Metric, error) {
	metricReader := NewMetricReader(reader)
	pointBuilder := NewPointBuilder(src, filteredPoints)
	var points []wf.Metric
	var err error
	for !metricReader.Done() {
		var parser expfmt.TextParser
		reader := bytes.NewReader(metricReader.Read())
		metricFamilies, err := parser.TextToMetricFamilies(reader)
		if err != nil {
			log.Errorf("reading text format failed: %s", err)
		}
		// TODO bug: err is overwritten here and above for every metric,
		// so whatever happens to be the last value of err is what is returned
		pointsToAdd, err := pointBuilder.build(metricFamilies)
		points = append(points, pointsToAdd...)
	}
	return points, err
}

type prometheusProvider struct {
	metrics.DefaultSourceProvider
	name              string
	useLeaderElection bool
	sources           []metrics.Source
}

func (p *prometheusProvider) GetMetricsSources() []metrics.Source {
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

func NewPrometheusProvider(cfg configuration.PrometheusSourceConfig) (metrics.SourceProvider, error) {
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
	filters := filter.FromConfig(cfg.Filters) // TODO test all allow and denylist stuff?

	var sources []metrics.Source
	metricsSource, err := NewPrometheusMetricsSource(cfg.URL, prefix, source, discovered, tags, filters, httpCfg)
	if err == nil {
		sources = append(sources, metricsSource)
	} else {
		return nil, fmt.Errorf("error creating source: %v", err)
	}

	return &prometheusProvider{
		name:              name,
		useLeaderElection: cfg.UseLeaderElection || discovered == "",
		sources:           sources,
	}, nil
}
