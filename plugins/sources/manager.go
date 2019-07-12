// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// enhanced to support multiple sources

package sources

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/prometheus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/stats"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/summary"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/systemd"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/telegraf"

	"github.com/golang/glog"
	gometrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
)

const (
	jitterMs = 4
)

var (
	providerCount  gometrics.Gauge
	sourceCount    gometrics.Gauge
	scrapeTimeout  gometrics.Gauge
	scrapeErrors   gometrics.Counter
	scrapeTimeouts gometrics.Counter
	scrapeLatency  gometrics.Histogram
)

func init() {
	providerCount = gometrics.GetOrRegisterGauge("source.manager.providers", gometrics.DefaultRegistry)
	sourceCount = gometrics.GetOrRegisterGauge("source.manager.sources", gometrics.DefaultRegistry)
	scrapeErrors = gometrics.GetOrRegisterCounter("source.manager.scrape.errors", gometrics.DefaultRegistry)
	scrapeTimeouts = gometrics.GetOrRegisterCounter("source.manager.scrape.timeouts", gometrics.DefaultRegistry)
	scrapeLatency = reporting.NewHistogram()
	_ = gometrics.Register("source.manager.scrape.latency", scrapeLatency)
}

// SourceManager ProviderHandler with gometrics gatherin support
type SourceManager interface {
	metrics.ProviderHandler
	GetPendingMetrics() []*metrics.DataBatch
}

type sourceManagerImpl struct {
	mtx                      sync.RWMutex
	gometricsSourceProviders map[string]metrics.MetricsSourceProvider
	responseChannel          chan *metrics.DataBatch
	response                 []*metrics.DataBatch
	responseMtx              sync.Mutex
}

func newEmptySourceManager() SourceManager {
	sm := &sourceManagerImpl{
		responseChannel:          make(chan *metrics.DataBatch),
		gometricsSourceProviders: make(map[string]metrics.MetricsSourceProvider),
	}

	sm.rotateResponse()
	go sm.run()

	return sm
}

// NewSourceManager creates a new NewSourceManager with the configured goMetricsSourceProviders
func NewSourceManager(src flags.Uris, statsPrefix string) SourceManager {
	sm := newEmptySourceManager()

	gometricsSourceProviders := buildProviders(src, statsPrefix)
	providerCount.Update(int64(len(gometricsSourceProviders)))
	for _, runtime := range gometricsSourceProviders {
		sm.AddProvider(runtime)
	}

	return sm
}

// AddProvider register and start a new goMetricsSourceProvider
func (sm *sourceManagerImpl) AddProvider(provider metrics.MetricsSourceProvider) {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()
	sm.gometricsSourceProviders[provider.Name()] = provider
	glog.V(4).Infof("added provider: %s", provider.Name())

	for _, source := range provider.GetMetricsSources() { // TODO: move loop to the scrape?
		go scrape(source, sm.responseChannel, provider.TimeOut())
	}

	ticker := time.NewTicker(provider.CollectionInterval())
	go func() {
		for range ticker.C {
			for _, source := range provider.GetMetricsSources() {
				go scrape(source, sm.responseChannel, provider.TimeOut())
			}
		}
	}()
}

func (sm *sourceManagerImpl) DeleteProvider(name string) {
	sm.mtx.Lock()
	defer sm.mtx.Unlock()
	delete(sm.gometricsSourceProviders, name)
	glog.V(4).Infof("deleted provider %s", name)
}

func (sm *sourceManagerImpl) run() {
	for {
		dataBatch := <-sm.responseChannel
		if dataBatch != nil {
			sm.responseMtx.Lock()
			sm.response = append(sm.response, dataBatch)
			sm.responseMtx.Unlock()
		}
	}
}

func (sm *sourceManagerImpl) rotateResponse() []*metrics.DataBatch {
	sm.responseMtx.Lock()
	defer sm.responseMtx.Unlock()
	response := sm.response
	sm.response = make([]*metrics.DataBatch, 0)
	return response
}

func scrape(source metrics.MetricsSource, channel chan *metrics.DataBatch, timeout time.Duration) {
	// Prevents network congestion.
	jitter := time.Duration(rand.Intn(jitterMs)) * time.Millisecond
	time.Sleep(jitter)

	scrapeStart := time.Now()
	timeoutTime := scrapeStart.Add(timeout)

	glog.V(2).Infof("Querying source: '%s'", source.Name())
	gometrics, err := source.ScrapeMetrics()
	if err != nil {
		scrapeErrors.Inc(1)
		glog.Errorf("Error in scraping containers from '%s': %v", source.Name(), err)
		return
	}

	now := time.Now()
	latency := now.Sub(scrapeStart)
	scrapeLatency.Update(latency.Nanoseconds())

	if !now.Before(timeoutTime) {
		scrapeTimeouts.Inc(1)
		glog.Warningf("Failed to get '%s' response in time", source.Name())
		return
	}
	channel <- gometrics
	glog.V(2).Infof("Done Querying source: '%s' (%v metrics) (%v latency)", source.Name(), len(gometrics.MetricPoints), latency)
}

func (sm *sourceManagerImpl) GetPendingMetrics() []*metrics.DataBatch {
	response := sm.rotateResponse()
	return response
}

func buildProviders(uris flags.Uris, statsPrefix string) []metrics.MetricsSourceProvider {
	result := make([]metrics.MetricsSourceProvider, 0, len(uris))
	for _, uri := range uris {
		provider, err := buildProvider(uri)
		if err == nil {
			result = append(result, provider)
		} else {
			glog.Errorf("Failed to create %v source: %v", uri, err)
		}
	}

	if len([]flags.Uri(uris)) != 0 && len(result) == 0 {
		glog.Fatal("No available source to use")
	}

	provider, _ := stats.NewInternalStatsProvider(statsPrefix)
	result = append(result, provider)
	return result
}

func buildProvider(uri flags.Uri) (metrics.MetricsSourceProvider, error) {
	var provider metrics.MetricsSourceProvider
	var err error

	switch uri.Key {
	case "kubernetes.summary_api":
		provider, err = summary.NewSummaryProvider(&uri.Val)
	case "prometheus":
		provider, err = prometheus.NewPrometheusProvider(&uri.Val)
	case "telegraf":
		provider, err = telegraf.NewProvider(&uri.Val)
	case "systemd":
		provider, err = systemd.NewProvider(&uri.Val)
	default:
		err = fmt.Errorf("source not recognized: %s", uri.Key)
	}

	if err == nil {
		if i, ok := provider.(metrics.ConfigurabeMetricsSourceProvider); ok {
			i.Configure(&uri.Val)
		}
	}

	return provider, err
}
