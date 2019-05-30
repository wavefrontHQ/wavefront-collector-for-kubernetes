// Based on https://github.com/kubernetes-retired/heapster/blob/master/metrics/sources/manager.go
// Diff against master for changes to the original code.

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
	"math/rand"
	"sync"
	"time"

	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"github.com/golang/glog"
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront/reporting"
)

const (
	MaxDelayMs       = 4 * 1000
	DelayPerSourceMs = 8
)

var (
	providerCount  metrics.Gauge
	sourceCount    metrics.Gauge
	scrapeTimeout  metrics.Gauge
	scrapeErrors   metrics.Counter
	scrapeTimeouts metrics.Counter
	scrapeLatency  metrics.Histogram
)

func init() {
	providerCount = metrics.GetOrRegisterGauge("source.manager.providers", metrics.DefaultRegistry)
	sourceCount = metrics.GetOrRegisterGauge("source.manager.sources", metrics.DefaultRegistry)
	scrapeErrors = metrics.GetOrRegisterCounter("source.manager.scrape.errors", metrics.DefaultRegistry)
	scrapeTimeouts = metrics.GetOrRegisterCounter("source.manager.scrape.timeouts", metrics.DefaultRegistry)
	scrapeLatency = reporting.NewHistogram()
	_ = metrics.Register("source.manager.scrape.latency", scrapeLatency)
}

func NewSourceManager(metricsSourceProviders []MetricsSourceProvider, metricsScrapeTimeout time.Duration) (MetricsSource, error) {
	providers := make(map[string]MetricsSourceProvider)
	for _, provider := range metricsSourceProviders {
		providers[provider.Name()] = provider
	}
	providerCount.Update(int64(len(providers)))
	return &sourceManager{
		metricsSourceProviders: providers,
		metricsScrapeTimeout:   metricsScrapeTimeout,
	}, nil
}

type sourceManager struct {
	metricsScrapeTimeout   time.Duration
	mtx                    sync.RWMutex
	metricsSourceProviders map[string]MetricsSourceProvider
}

func (this *sourceManager) Name() string {
	return "source_manager"
}

func (this *sourceManager) AddProvider(provider MetricsSourceProvider) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.metricsSourceProviders[provider.Name()] = provider
	glog.V(4).Infof("added provider: %s", provider.Name())
}

func (this *sourceManager) DeleteProvider(name string) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	delete(this.metricsSourceProviders, name)
	glog.V(4).Infof("deleted provider %s", name)
}

func (this *sourceManager) ScrapeMetrics(start, end time.Time) (*DataBatch, error) {
	glog.V(1).Infof("Scraping metrics start: %s, end: %s", start, end)

	sources := []MetricsSource{}
	this.mtx.RLock()
	for _, sourceProvider := range this.metricsSourceProviders {
		glog.V(2).Infof("Scraping sources from provider: %s", sourceProvider.Name())
		sources = append(sources, sourceProvider.GetMetricsSources()...)
	}
	providerCount.Update(int64(len(this.metricsSourceProviders)))
	this.mtx.RUnlock()

	sourceCount.Update(int64(len(sources)))

	responseChannel := make(chan *DataBatch)
	startTime := time.Now()
	timeoutTime := startTime.Add(this.metricsScrapeTimeout)

	delayMs := DelayPerSourceMs * len(sources)
	if delayMs > MaxDelayMs {
		delayMs = MaxDelayMs
	}

	for _, source := range sources {

		go func(source MetricsSource, channel chan *DataBatch, start, end, timeoutTime, scrapeStart time.Time, delayInMs int) {

			// Prevents network congestion.
			time.Sleep(time.Duration(rand.Intn(delayMs)) * time.Millisecond)

			glog.V(2).Infof("Querying source: %s", source.Name())
			metrics, err := scrape(source, start, end)
			if err != nil {
				scrapeErrors.Inc(1)
				glog.Errorf("Error in scraping containers from %s: %v", source.Name(), err)
				return
			}

			now := time.Now()
			latency := now.Sub(scrapeStart).Nanoseconds()
			scrapeLatency.Update(latency)

			if !now.Before(timeoutTime) {
				scrapeTimeouts.Inc(1)
				glog.Warningf("Failed to get %s response in time", source)
				return
			}
			timeForResponse := timeoutTime.Sub(now)

			select {
			case channel <- metrics:
				// passed the response correctly.
				return
			case <-time.After(timeForResponse):
				scrapeTimeouts.Inc(1)
				glog.Warningf("Failed to send the response back %s", source)
				return
			}
		}(source, responseChannel, start, end, timeoutTime, startTime, delayMs)
	}
	response := DataBatch{
		Timestamp:  end,
		MetricSets: map[string]*MetricSet{},
	}

responseloop:
	for i := range sources {
		now := time.Now()
		if !now.Before(timeoutTime) {
			glog.Warningf("Failed to get all responses in time (got %d/%d)", i, len(sources))
			break
		}

		select {
		case dataBatch := <-responseChannel:
			if dataBatch != nil {
				for key, value := range dataBatch.MetricSets {
					response.MetricSets[key] = value
				}
				if len(dataBatch.MetricPoints) > 0 {
					response.MetricPoints = append(response.MetricPoints, dataBatch.MetricPoints...)
				}
			}

		case <-time.After(timeoutTime.Sub(now)):
			glog.Warningf("Failed to get all responses in time (got %d/%d)", i, len(sources))
			break responseloop
		}
	}
	glog.V(1).Infof("ScrapeMetrics: time: %s size: %d", time.Since(startTime), len(response.MetricSets))
	return &response, nil
}

func scrape(s MetricsSource, start, end time.Time) (*DataBatch, error) {
	return s.ScrapeMetrics(start, end)
}
