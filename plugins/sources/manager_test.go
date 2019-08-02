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

package sources

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
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
		for _, point := range dataBatch.MetricPoints {
			present[point.Metric] = true
		}
	}

	Manager().StopProviders()

	assert.True(t, present["nto_1"], "nto_1 not found - present:%v", present)
	assert.True(t, present["nto_2"], "nto_2 not found - present:%v", present)
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
		for _, point := range dataBatch.MetricPoints {
			present[point.Metric] = true
		}
	}

	Manager().StopProviders()

	assert.True(t, present["s1"], "s1 not found - present:%v", present)
	assert.False(t, present["s2"], "s2 found - present:%v", present)

}

func TestMultipleMetrics(t *testing.T) {
	msp1 := util.NewDummyMetricsSourceProvider(
		"p1", 10*time.Millisecond, 10*time.Millisecond,
		util.NewDummyMetricsSource("s1", 0))

	msp2 := util.NewDummyMetricsSourceProvider(
		"p2", 10*time.Millisecond, 10*time.Millisecond,
		util.NewDummyMetricsSource("s2", 0))

	Manager().AddProvider(msp1)
	Manager().AddProvider(msp2)

	time.Sleep(105 * time.Millisecond)
	Manager().DeleteProvider("p2")
	time.Sleep(95 * time.Millisecond)
	Manager().DeleteProvider("p1")

	dataBatchList := Manager().GetPendingMetrics()

	counts := make(map[string]int)
	for _, dataBatch := range dataBatchList {
		for _, point := range dataBatch.MetricPoints {
			counts[point.Metric]++
		}
	}

	Manager().StopProviders()

	assert.True(t, (counts["s1"] >= 20) && (counts["s1"] <= 21), counts["s1"], "incorrect s1 count - counts: %vs", counts)
	assert.Equal(t, 10, counts["s2"], "incorrect s2 count - counts: %v", counts)
}

func TestConfig(t *testing.T) {
	var provider metrics.MetricsSourceProvider

	provider = &testProvider{}

	if i, ok := provider.(metrics.ConfigurabeMetricsSourceProvider); ok {
		i.Configure(time.Hour*1, time.Minute*1)
		log.Infof("Name: %s - CollectionInterval: %v", provider.Name(), provider.CollectionInterval())
	}
	assert.Equal(t, time.Hour, provider.CollectionInterval(), "incorrect CollectionInterval")
	assert.Equal(t, time.Minute, provider.Timeout(), "incorrect Timeout")
}

type testProvider struct {
	metrics.DefaultMetricsSourceProvider
}

func (p *testProvider) GetMetricsSources() []metrics.MetricsSource {
	return make([]metrics.MetricsSource, 0)
}

func (p *testProvider) Name() string {
	return "testProvider"
}
