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

package manager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources"
)

func TestFlow(t *testing.T) {
	provider := util.NewDummyMetricsSourceProvider(
		"p1", 100*time.Millisecond, 100*time.Millisecond,
		util.NewDummyMetricsSource("src", time.Millisecond))

	sink := util.NewDummySink("sink", time.Millisecond)
	processor := util.NewDummyDataProcessor(time.Millisecond)

	sources.Manager().AddProvider(provider)

	manager, _ := NewFlushManager([]metrics.DataProcessor{processor}, sink, 100*time.Millisecond)
	manager.Start()

	// 4-5 cycles
	time.Sleep(time.Millisecond * 550)
	manager.Stop()

	if sink.GetExportCount() < 4 || sink.GetExportCount() > 5 {
		t.Fatalf("Wrong number of exports executed: %d", sink.GetExportCount())
	}
}

func TestCombineMetricSets(t *testing.T) {
	dst := &metrics.DataBatch{}
	assert.Nil(t, dst.MetricSets)

	firstBatch := createDataBatch("node_1")
	combineMetricSets(firstBatch, dst)
	assert.Equal(t, 4, len(dst.MetricSets))
	testKeysAndValues(t, firstBatch, dst)

	secondBatch := createDataBatch("node_2")
	combineMetricSets(secondBatch, dst)
	assert.Equal(t, 8, len(dst.MetricSets))
	testKeysAndValues(t, secondBatch, dst)
}

func testKeysAndValues(t *testing.T, src, dst *metrics.DataBatch) {
	for k, v := range src.MetricSets {
		if dstVal, found := dst.MetricSets[k]; found {
			assert.Equal(t, v, dstVal)
		} else {
			assert.Fail(t, "failed to find metric set: %s", k)
		}
	}
}

func createDataBatch(prefix string) *metrics.DataBatch {
	batch := metrics.DataBatch{
		Timestamp:  time.Now(),
		MetricSets: map[string]*metrics.MetricSet{},
	}
	batch.MetricSets[prefix+"m1"] = createMetricSet("cpu/limit", metrics.MetricGauge, 1000)
	batch.MetricSets[prefix+"m2"] = createMetricSet("cpu/usage", metrics.MetricCumulative, 43363664)
	batch.MetricSets[prefix+"m3"] = createMetricSet("memory/limit", metrics.MetricGauge, -1)
	batch.MetricSets[prefix+"m4"] = createMetricSet("memory/usage", metrics.MetricGauge, 487424)
	return &batch
}

func createMetricSet(name string, metricType metrics.MetricType, value int64) *metrics.MetricSet {
	set := &metrics.MetricSet{
		MetricValues: map[string]metrics.MetricValue{
			name: {
				ValueType: metrics.ValueInt64,
				IntValue:  value,
			},
		},
	}
	return set
}
