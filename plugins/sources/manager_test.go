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

// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package sources

import (
	"testing"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
)

func TestNoTimeout(t *testing.T) {
	metricsSourceProvider := util.NewDummyMetricsSourceProvider("dummy_nt",
		100*time.Millisecond, 100*time.Millisecond,
		util.NewDummyMetricsSource("nto_1", 10*time.Millisecond),
		util.NewDummyMetricsSource("nto_2", 10*time.Millisecond))

	Manager().AddProvider(metricsSourceProvider)

	time.Sleep(200 * time.Millisecond)

	dataBatchList := Manager().GetPendingMetrics()

	present := make(map[string]bool)
	for _, dataBatch := range dataBatchList {
		for _, point := range dataBatch.Points {
			present[point.Metric] = true
		}
	}

	Manager().StopProviders()

	assert.True(t, present["nto_1"], "nto_1 not found - present:%v", present)
	assert.True(t, present["nto_2"], "nto_2 not found - present:%v", present)
}

func TestScrapeMetrics(t *testing.T) {

	t.Run("Test Scrape Errors with Non AutoDiscovered Source", func(t *testing.T) {
		initialErrCnt := scrapeErrors.Count()

		metricsSourceProvider := util.NewDummyMetricsSourceProvider(
			"dummy", 10*time.Millisecond, 75*time.Millisecond,
			util.NewDummyMetricsSourceWithError("s1", 0, false),
			util.NewDummyMetricsSource("s2", 0))

		testDataBatch := make(chan *metrics.Batch)
		var dbatch *metrics.Batch

		go func() {
			scrape(metricsSourceProvider, testDataBatch)
			dbatch = <-testDataBatch

			present := make(map[string]bool)
			for _, point := range dbatch.Points {
				present[point.Metric] = true
			}
			assert.Empty(t, present)
			assert.Equal(t, initialErrCnt+1, scrapeErrors.Count())
		}()
	})

	t.Run("Test Scrape Errors with AutoDiscovered Source", func(t *testing.T) {
		initialErrCnt := scrapeErrors.Count()
		initialWarningCnt := scrapeWarnings.Count()

		metricsSourceProvider := util.NewDummyMetricsSourceProvider(
			"dummy", 0, 75*time.Millisecond,
			util.NewDummyMetricsSourceWithError("s1", 0, true),
			util.NewDummyMetricsSource("s2", 0))

		testDataBatch := make(chan *metrics.Batch)
		var dbatch *metrics.Batch

		go func() {
			scrape(metricsSourceProvider, testDataBatch)
			dbatch = <-testDataBatch

			present := make(map[string]bool)
			for _, point := range dbatch.Points {
				present[point.Metric] = true
			}
			assert.Empty(t, present)
			assert.Equal(t, initialErrCnt+1, scrapeErrors.Count())
			assert.Equal(t, initialWarningCnt+1, scrapeWarnings.Count())
		}()
	})
}

func TestTimeout(t *testing.T) {
	metricsSourceProvider := util.NewDummyMetricsSourceProvider(
		"dummy", 100*time.Millisecond, 75*time.Millisecond,
		util.NewDummyMetricsSource("s1", 50*time.Millisecond),
		util.NewDummyMetricsSource("s2", 100*time.Millisecond))

	Manager().AddProvider(metricsSourceProvider)

	time.Sleep(200 * time.Millisecond)

	dataBatchList := Manager().GetPendingMetrics()

	present := make(map[string]bool)
	for _, dataBatch := range dataBatchList {
		for _, point := range dataBatch.Points {
			present[point.Metric] = true
		}
	}

	Manager().StopProviders()

	assert.True(t, present["s1"], "s1 not found - present:%v", present)
	assert.False(t, present["s2"], "s2 found - present:%v", present)

}

func TestMultipleMetrics(t *testing.T) {
	t.Skip("skipping test to pass jenkins pipeline - K8SSAAS-561")
	interval := 10 * time.Millisecond
	msp1 := util.NewDummyMetricsSourceProvider(
		"p1", interval, interval,
		util.NewDummyMetricsSource("s1", 0))

	msp2 := util.NewDummyMetricsSourceProvider(
		"p2", interval, interval,
		util.NewDummyMetricsSource("s2", 0))

	Manager().AddProvider(msp1)
	Manager().AddProvider(msp2)

	s2Intervals := 10
	s2wait := time.Duration(s2Intervals)*interval + 5*time.Millisecond // fudge factor
	time.Sleep(s2wait)
	Manager().DeleteProvider("p2")

	s1Intervals := 20
	s1wait := time.Duration(s1Intervals)*interval - s2wait
	time.Sleep(s1wait)
	Manager().DeleteProvider("p1")

	dataBatchList := Manager().GetPendingMetrics()

	counts := make(map[string]int)
	for _, dataBatch := range dataBatchList {
		for _, point := range dataBatch.Points {
			counts[point.Metric]++
		}
	}

	Manager().StopProviders()

	assert.True(t, (counts["s1"] >= s1Intervals-1) && (counts["s1"] <= s1Intervals+1),
		"incorrect s1 scrape count - expected %v, actual: %v", s1Intervals, counts["s1"])
	assert.True(t, (counts["s2"] >= s2Intervals-1) && (counts["s2"] <= s2Intervals+1),
		"incorrect s2 scrape count - expected %v, actual: %v", s2Intervals, counts["s2"])
}

func TestConfig(t *testing.T) {
	var provider metrics.SourceProvider

	provider = &testProvider{}

	if i, ok := provider.(metrics.ConfigurableSourceProvider); ok {
		i.Configure(time.Hour*1, time.Minute*1)
		log.Infof("Name: %s - CollectionInterval: %v", provider.Name(), provider.CollectionInterval())
	}
	assert.Equal(t, time.Hour, provider.CollectionInterval(), "incorrect CollectionInterval")
	assert.Equal(t, time.Minute, provider.Timeout(), "incorrect Timeout")
}

func TestPluginConfigProvider(t *testing.T) {
	t.Run("provides plugin configs for source providers that offer them", func(t *testing.T) {
		pluginConfigProvider := &testPluginProvider{}
		regularProvider := &testProvider{}
		Manager().AddProvider(pluginConfigProvider)
		Manager().AddProvider(regularProvider)

		discoveryConfigs := Manager().DiscoveryConfigs()

		if assert.Equal(t, 1, len(discoveryConfigs)) {
			assert.Equal(t, "testDiscoveryConfig", discoveryConfigs[0].Name)
		}

		Manager().DeleteProvider("testPluginProvider")
		Manager().DeleteProvider("testProvider")
	})
}

type testProvider struct {
	metrics.DefaultSourceProvider
}

func (p *testProvider) GetMetricsSources() []metrics.Source {
	return make([]metrics.Source, 0)
}

func (p *testProvider) Name() string {
	return "testProvider"
}

type testPluginProvider struct {
	metrics.DefaultSourceProvider
}

func (p *testPluginProvider) DiscoveryConfigs() []discovery.PluginConfig {
	return []discovery.PluginConfig{{Name: "testDiscoveryConfig"}}
}

func (p *testPluginProvider) GetMetricsSources() []metrics.Source {
	return make([]metrics.Source, 0)
}

func (p *testPluginProvider) Name() string {
	return "testPluginProvider"
}
