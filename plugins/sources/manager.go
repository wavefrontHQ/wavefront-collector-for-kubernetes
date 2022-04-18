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
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"math/rand"
	"sort"
	"sync"
	"time"

    "github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/controlplane"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/cadvisor"

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
	scrapeWarnings gometrics.Counter
	scrapeTimeouts gometrics.Counter
	scrapesMissed  gometrics.Counter
	scrapeLatency  gometrics.Histogram
	singleton      *sourceManagerImpl
	once           sync.Once
)

func init() {
	providerCount = gometrics.GetOrRegisterGauge("source.manager.providers", gometrics.DefaultRegistry)
	scrapeErrors = gometrics.GetOrRegisterCounter("source.manager.scrape.errors", gometrics.DefaultRegistry)
	scrapeWarnings = gometrics.GetOrRegisterCounter("source.manager.scrape.warnings", gometrics.DefaultRegistry)
	scrapeTimeouts = gometrics.GetOrRegisterCounter("source.manager.scrape.timeouts", gometrics.DefaultRegistry)
	scrapesMissed = gometrics.GetOrRegisterCounter("source.manager.scrape.missed", gometrics.DefaultRegistry)
	scrapeLatency = reporting.NewHistogram()
	_ = gometrics.Register("source.manager.scrape.latency", scrapeLatency)
}

// SourceManager ProviderHandler with metrics gathering support
type SourceManager interface {
	metrics.ProviderHandler
    discovery.ConfigProvider

	StopProviders()
	GetPendingMetrics() []*metrics.Batch
	SetDefaultCollectionInterval(time.Duration)
	BuildProviders(config configuration.SourceConfig) error
}

type sourceManagerImpl struct {
	responseChannel           chan *metrics.Batch
	defaultCollectionInterval time.Duration

	metricsSourcesMtx      sync.Mutex
	metricsSourceProviders map[string]metrics.SourceProvider
	metricsSourceTimers    map[string]*IntervalTimer
	metricsSourceQuits     map[string]chan struct{}

	responseMtx sync.Mutex
	response    []*metrics.Batch
}

// Manager return the SourceManager
func Manager() SourceManager {
	once.Do(func() {
		singleton = &sourceManagerImpl{
			responseChannel:           make(chan *metrics.Batch),
			metricsSourceProviders:    make(map[string]metrics.SourceProvider),
			metricsSourceTimers:       make(map[string]*IntervalTimer),
			metricsSourceQuits:        make(map[string]chan struct{}),
			defaultCollectionInterval: time.Minute,
		}
		singleton.rotateResponse()
		go singleton.run()
	})
	return singleton
}

func (sm *sourceManagerImpl) DiscoveryConfigs() []discovery.PluginConfig {
    var pluginConfigs []discovery.PluginConfig
    for _, provider := range sm.metricsSourceProviders{
        if pluginProvider, ok := provider.(discovery.ConfigProvider); ok {
            pluginConfigs = append(pluginConfigs, pluginProvider.DiscoveryConfigs()...)
        }
    }
    return pluginConfigs
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

// AddProvider register and start a new SourceProvider
func (sm *sourceManagerImpl) AddProvider(provider metrics.SourceProvider) {
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

	var interval time.Duration
	if provider.CollectionInterval() > 0 {
		interval = provider.CollectionInterval()
	} else {
		interval = sm.defaultCollectionInterval

		log.WithFields(log.Fields{
			"provider":            name,
			"collection_interval": sm.defaultCollectionInterval,
		}).Info("Using default collection interval")
	}
	intervalTimer := NewIntervalTimer(interval)

	quit := make(chan struct{})

	sm.metricsSourceProviders[name] = provider
	sm.metricsSourceTimers[name] = intervalTimer
	sm.metricsSourceQuits[name] = quit

	providerCount.Update(int64(len(sm.metricsSourceProviders)))

	go func() {
		for {
			select {
			case <-intervalTimer.C:
				scrape(provider, sm.responseChannel)
				scrapesMissed.Inc(intervalTimer.Reset())
			case <-quit:
				return
			}
		}
	}()
}

func (sm *sourceManagerImpl) DeleteProvider(name string) {
	provider, found := sm.metricsSourceProviders[name]
	if !found {
		log.Debugf("Metrics Source Provider '%s' not found", name)
		return
	}

	sm.metricsSourcesMtx.Lock()
	defer sm.metricsSourcesMtx.Unlock()

	delete(sm.metricsSourceProviders, name)
	if ticker, ok := sm.metricsSourceTimers[name]; ok {
		ticker.Stop()
		delete(sm.metricsSourceTimers, name)
	}
	if quit, ok := sm.metricsSourceQuits[name]; ok {
		close(quit)
		delete(sm.metricsSourceQuits, name)
	}

	for _, source := range provider.GetMetricsSources() {
		source.Cleanup()
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

func (sm *sourceManagerImpl) rotateResponse() []*metrics.Batch {
	sm.responseMtx.Lock()
	defer sm.responseMtx.Unlock()
	response := sm.response
	sm.response = make([]*metrics.Batch, 0)
	return response
}

func scrape(provider metrics.SourceProvider, channel chan *metrics.Batch) {
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

		dataBatch, err := source.Scrape()
		if err != nil {
			if source.AutoDiscovered() {
				log.Warningf("Could not scrape containers, skipping source '%s': %v", source.Name(), err)
				scrapeWarnings.Inc(1)
			} else {
				log.Errorf("Error in scraping containers from '%s': %v", source.Name(), err)
				scrapeErrors.Inc(1)
			}
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
			"total_metrics": len(dataBatch.Points) + len(dataBatch.Sets),
			"latency":       latency,
		}).Infof("Finished querying source")
	}
}

func (sm *sourceManagerImpl) GetPendingMetrics() []*metrics.Batch {
	response := sm.rotateResponse()
	sort.Slice(response, func(i, j int) bool { return response[i].Timestamp.Before(response[j].Timestamp) })
	return response
}

func buildProviders(cfg configuration.SourceConfig) (result []metrics.SourceProvider) {
	if cfg.SummaryConfig != nil {
		provider, err := summary.NewSummaryProvider(*cfg.SummaryConfig)
		result = appendProvider(result, provider, err, cfg.SummaryConfig.Collection)

		if cfg.CadvisorConfig != nil {
			provider, err = cadvisor.NewProvider(*cfg.CadvisorConfig, *cfg.SummaryConfig)
			result = appendProvider(result, provider, err, cfg.CadvisorConfig.Collection)
		}
		if cfg.ControlPlaneConfig != nil {
			provider, err = controlplane.NewProvider(*cfg.ControlPlaneConfig, *cfg.SummaryConfig)
            result = appendProvider(result, provider, err, cfg.ControlPlaneConfig.Collection)
		}
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

func appendProvider(
	slice []metrics.SourceProvider,
	provider metrics.SourceProvider,
	err error,
	cfg configuration.CollectionConfig,
) []metrics.SourceProvider {

	if err != nil {
		log.Errorf("Failed to create source: %v", err)
		return slice
	}
	slice = append(slice, provider)
	if i, ok := provider.(metrics.ConfigurableSourceProvider); ok {
		i.Configure(cfg.Interval, cfg.Timeout)
	}
	return slice
}
