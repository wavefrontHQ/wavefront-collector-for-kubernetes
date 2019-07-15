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
	Stop()
}

type sourceManagerImpl struct {
	metricsSourcesMtx      sync.Mutex
	metricsSourceProviders map[string]metrics.MetricsSourceProvider
	metricsSourceTickers   map[string]*time.Ticker
	responseChannel        chan *metrics.DataBatch
	response               []*metrics.DataBatch
	responseMtx            sync.Mutex
}

func NewEmptySourceManager() SourceManager {
	sm := &sourceManagerImpl{
		responseChannel:        make(chan *metrics.DataBatch),
		metricsSourceProviders: make(map[string]metrics.MetricsSourceProvider),
		metricsSourceTickers:   make(map[string]*time.Ticker),
	}

	sm.rotateResponse()
	go sm.run()

	return sm
}

// NewSourceManager creates a new NewSourceManager with the configured goMetricsSourceProviders
func NewSourceManager(src flags.Uris) SourceManager {
	sm := NewEmptySourceManager()

	gometricsSourceProviders := buildProviders(src)
	providerCount.Update(int64(len(gometricsSourceProviders)))
	for _, runtime := range gometricsSourceProviders {
		sm.AddProvider(runtime)
	}

	return sm
}

// AddProvider register and start a new goMetricsSourceProvider
func (sm *sourceManagerImpl) AddProvider(provider metrics.MetricsSourceProvider) {
	name := provider.Name()
	glog.Infof("Adding provider: '%s' - collection iterval: '%v' - timeout: '%v'", name, provider.CollectionInterval(), provider.TimeOut())
	if _, found := sm.metricsSourceProviders[name]; found {
		glog.Fatalf("Error on 'SourceManager.AddProvider' Duplicate Metrics Source Provider name: '%s'", name)
	}

	sm.metricsSourcesMtx.Lock()
	defer sm.metricsSourcesMtx.Unlock()

	ticker := time.NewTicker(provider.CollectionInterval())

	sm.metricsSourceProviders[name] = provider
	sm.metricsSourceTickers[name] = ticker
	glog.V(2).Infof("added provider: %s", name)

	go scrape(provider, sm.responseChannel)
	go func() {
		for range ticker.C {
			go scrape(provider, sm.responseChannel)
		}
	}()
}

func (sm *sourceManagerImpl) DeleteProvider(name string) {
	if _, found := sm.metricsSourceProviders[name]; !found {
		glog.Fatalf("Error on 'SourceManager.DeleteProvider'  Metrics Source Provider '%s' not found", name)
	}

	sm.metricsSourcesMtx.Lock()
	defer sm.metricsSourcesMtx.Unlock()

	delete(sm.metricsSourceProviders, name)
	if tiker, ok := sm.metricsSourceTickers[name]; ok {
		tiker.Stop()
		delete(sm.metricsSourceTickers, name)
	}
	glog.Infof("deleted provider %s", name)
}

func (sm *sourceManagerImpl) Stop() {
	for provider := range sm.metricsSourceProviders {
		sm.DeleteProvider(provider)
	}
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

func scrape(provider metrics.MetricsSourceProvider, channel chan *metrics.DataBatch) {
	for _, source := range provider.GetMetricsSources() {
		// Prevents network congestion.
		jitter := time.Duration(rand.Intn(jitterMs)) * time.Millisecond
		time.Sleep(jitter)

		scrapeStart := time.Now()
		timeoutTime := scrapeStart.Add(provider.TimeOut())

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
}

func (sm *sourceManagerImpl) GetPendingMetrics() []*metrics.DataBatch {
	response := sm.rotateResponse()
	return response
}

func buildProviders(uris flags.Uris) []metrics.MetricsSourceProvider {
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
	case "internal_stats":
		provider, err = stats.NewInternalStatsProvider(&uri.Val)
	default:
		err = fmt.Errorf("source not recognized: %s", uri.Key)
	}

	if err == nil {
		if i, ok := provider.(metrics.ConfigurabeMetricsSourceProvider); ok {
			i.Configure(&uri.Val)
		} else {
			if provider.CollectionInterval() == 0 {
				return nil, fmt.Errorf("Provider '%s' have a invalid 'CollectionInterval'", provider.Name())
			}
		}
	}

	return provider, err
}
