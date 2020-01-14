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

// Copyright 2018-2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package sources

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/kstate"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/prometheus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/stats"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/summary"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/systemd"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/telegraf"

	gometrics "github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
)

const (
	jitterMs = 4
)

var (
	providerCount  gometrics.Gauge
	scrapeErrors   gometrics.Counter
	scrapeTimeouts gometrics.Counter
	scrapeLatency  gometrics.Histogram
	singleton      *sourceManagerImpl
	once           sync.Once
)

func init() {
	providerCount = gometrics.GetOrRegisterGauge("source.manager.providers", gometrics.DefaultRegistry)
	scrapeErrors = gometrics.GetOrRegisterCounter("source.manager.scrape.errors", gometrics.DefaultRegistry)
	scrapeTimeouts = gometrics.GetOrRegisterCounter("source.manager.scrape.timeouts", gometrics.DefaultRegistry)
	scrapeLatency = reporting.NewHistogram()
	_ = gometrics.Register("source.manager.scrape.latency", scrapeLatency)
}

// SourceManager ProviderHandler with metrics gathering support
type SourceManager interface {
	metrics.ProviderHandler

	StopProviders()
	GetPendingMetrics() []*metrics.DataBatch
	SetDefaultCollectionInterval(time.Duration)
	BuildProviders(config configuration.SourceConfig) error
}

type sourceManagerImpl struct {
	responseChannel           chan *metrics.DataBatch
	defaultCollectionInterval time.Duration

	metricsSourcesMtx      sync.Mutex
	metricsSourceProviders map[string]metrics.MetricsSourceProvider
	metricsSourceTickers   map[string]*time.Ticker
	metricsSourceQuits     map[string]chan struct{}

	responseMtx sync.Mutex
	response    []*metrics.DataBatch
}

// Manager return the SourceManager
func Manager() SourceManager {
	once.Do(func() {
		singleton = &sourceManagerImpl{
			responseChannel:           make(chan *metrics.DataBatch),
			metricsSourceProviders:    make(map[string]metrics.MetricsSourceProvider),
			metricsSourceTickers:      make(map[string]*time.Ticker),
			metricsSourceQuits:        make(map[string]chan struct{}),
			defaultCollectionInterval: time.Minute,
		}
		singleton.rotateResponse()
		go singleton.run()
	})
	return singleton
}

// BuildProviders creates a new source manager with the configured MetricsSourceProviders
func (sm *sourceManagerImpl) BuildProviders(cfg configuration.SourceConfig) error {
	sources := buildProviders(cfg)
	for _, runtime := range sources {
		sm.AddProvider(runtime)
	}
	if len(sm.metricsSourceProviders) == 0 {
		return fmt.Errorf("no available sources to use")
	}
	return nil
}

func (sm *sourceManagerImpl) SetDefaultCollectionInterval(defaultCollectionInterval time.Duration) {
	sm.defaultCollectionInterval = defaultCollectionInterval
}

// AddProvider register and start a new MetricsSourceProvider
func (sm *sourceManagerImpl) AddProvider(provider metrics.MetricsSourceProvider) {
	name := provider.Name()

	log.WithFields(log.Fields{
		"name":                name,
		"collection_interval": provider.CollectionInterval(),
		"timeout":             provider.Timeout(),
	}).Info("Adding provider")

	if _, found := sm.metricsSourceProviders[name]; found {
		log.WithField("name", name).Info("deleting existing provider")
		sm.DeleteProvider(name)
	}

	sm.metricsSourcesMtx.Lock()
	defer sm.metricsSourcesMtx.Unlock()

	var ticker *time.Ticker
	if provider.CollectionInterval() > 0 {
		ticker = time.NewTicker(provider.CollectionInterval())
	} else {
		ticker = time.NewTicker(sm.defaultCollectionInterval)

		log.WithFields(log.Fields{
			"provider":            name,
			"collection_interval": sm.defaultCollectionInterval,
		}).Info("Using default collection interval")
	}

	quit := make(chan struct{})

	sm.metricsSourceProviders[name] = provider
	sm.metricsSourceTickers[name] = ticker
	sm.metricsSourceQuits[name] = quit

	providerCount.Update(int64(len(sm.metricsSourceProviders)))

	go func() {
		for {
			select {
			case <-ticker.C:
				scrape(provider, sm.responseChannel)
			case <-quit:
				return
			}
		}
	}()
}

func (sm *sourceManagerImpl) DeleteProvider(name string) {
	if _, found := sm.metricsSourceProviders[name]; !found {
		log.Debugf("Metrics Source Provider '%s' not found", name)
		return
	}

	sm.metricsSourcesMtx.Lock()
	defer sm.metricsSourcesMtx.Unlock()

	delete(sm.metricsSourceProviders, name)
	if ticker, ok := sm.metricsSourceTickers[name]; ok {
		ticker.Stop()
		delete(sm.metricsSourceTickers, name)
	}
	if quit, ok := sm.metricsSourceQuits[name]; ok {
		close(quit)
		delete(sm.metricsSourceQuits, name)
	}
	log.WithField("name", name).Info("Deleted provider")
}

func (sm *sourceManagerImpl) StopProviders() {
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
		timeout := provider.Timeout()
		if timeout <= 0 {
			timeout = time.Minute
		}

		log.WithField("name", source.Name()).Info("Querying source")

		dataBatch, err := source.ScrapeMetrics()
		if err != nil {
			scrapeErrors.Inc(1)
			log.Errorf("Error in scraping containers from '%s': %v", source.Name(), err)
			return
		}

		now := time.Now()
		latency := now.Sub(scrapeStart)
		scrapeLatency.Update(latency.Nanoseconds())

		// always send the collected data even if latency > timeout
		channel <- dataBatch

		if !now.Before(scrapeStart.Add(timeout)) {
			scrapeTimeouts.Inc(1)
			log.Warningf("'%s' high response latency: %s", source.Name(), latency)
		}

		log.WithFields(log.Fields{
			"name":          source.Name(),
			"total_metrics": len(dataBatch.MetricPoints) + len(dataBatch.MetricSets),
			"latency":       latency,
		}).Debug("Finished querying source")
	}
}

func (sm *sourceManagerImpl) GetPendingMetrics() []*metrics.DataBatch {
	response := sm.rotateResponse()
	sort.Slice(response, func(i, j int) bool { return response[i].Timestamp.Before(response[j].Timestamp) })
	return response
}

func buildProviders(cfg configuration.SourceConfig) []metrics.MetricsSourceProvider {
	result := make([]metrics.MetricsSourceProvider, 0)

	if cfg.SummaryConfig != nil {
		provider, err := summary.NewSummaryProvider(*cfg.SummaryConfig)
		result = appendProvider(result, provider, err, cfg.SummaryConfig.Collection)
	}
	if cfg.SystemdConfig != nil {
		provider, err := systemd.NewProvider(*cfg.SystemdConfig)
		result = appendProvider(result, provider, err, cfg.SystemdConfig.Collection)
	}
	if cfg.StatsConfig != nil {
		provider, err := stats.NewInternalStatsProvider(*cfg.StatsConfig)
		result = appendProvider(result, provider, err, cfg.StatsConfig.Collection)
	}
	if cfg.StateConfig != nil {
		provider, err := kstate.NewStateProvider(*cfg.StateConfig)
		result = appendProvider(result, provider, err, cfg.StateConfig.Collection)
	}
	for _, srcCfg := range cfg.TelegrafConfigs {
		provider, err := telegraf.NewProvider(*srcCfg)
		result = appendProvider(result, provider, err, srcCfg.Collection)
	}
	for _, srcCfg := range cfg.PrometheusConfigs {
		provider, err := prometheus.NewPrometheusProvider(*srcCfg)
		result = appendProvider(result, provider, err, srcCfg.Collection)
	}

	if len(result) == 0 {
		log.Fatal("No available source to use")
	}
	return result
}

func appendProvider(slice []metrics.MetricsSourceProvider, provider metrics.MetricsSourceProvider, err error,
	cfg configuration.CollectionConfig) []metrics.MetricsSourceProvider {

	if err != nil {
		log.Errorf("Failed to create source: %v", err)
		return slice
	}
	slice = append(slice, provider)
	if err == nil {
		if i, ok := provider.(metrics.ConfigurabeMetricsSourceProvider); ok {
			i.Configure(cfg.Interval, cfg.Timeout)
		}
	}
	return slice
}
