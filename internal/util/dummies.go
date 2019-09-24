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

package util

import (
	"strings"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	v1 "k8s.io/api/core/v1"
)

type DummySink struct {
	name        string
	mutex       sync.Mutex
	exportCount int
	stopped     bool
	latency     time.Duration
}

func (dummy *DummySink) Name() string {
	return dummy.name
}
func (dummy *DummySink) ExportData(*DataBatch) {
	dummy.mutex.Lock()
	dummy.exportCount++
	dummy.mutex.Unlock()

	time.Sleep(dummy.latency)
}

func (dummy *DummySink) ExportEvents(function string, eNew *v1.Event, eOld *v1.Event) {
}

func (dummy *DummySink) Stop() {
	dummy.mutex.Lock()
	dummy.stopped = true
	dummy.mutex.Unlock()

	time.Sleep(dummy.latency)
}

func (dummy *DummySink) IsStopped() bool {
	dummy.mutex.Lock()
	defer dummy.mutex.Unlock()
	return dummy.stopped
}

func (dummy *DummySink) GetExportCount() int {
	dummy.mutex.Lock()
	defer dummy.mutex.Unlock()
	return dummy.exportCount
}

func NewDummySink(name string, latency time.Duration) *DummySink {
	return &DummySink{
		name:        name,
		latency:     latency,
		exportCount: 0,
		stopped:     false,
	}
}

type DummyMetricsSource struct {
	latency   time.Duration
	metricSet MetricSet
	name      string
}

func (dummy *DummyMetricsSource) Name() string {
	return dummy.name
}

func (dummy *DummyMetricsSource) ScrapeMetrics() (*DataBatch, error) {
	time.Sleep(dummy.latency)

	point := &metrics.MetricPoint{
		Metric:    strings.Replace(dummy.Name(), " ", ".", -1),
		Value:     1,
		Timestamp: time.Now().UnixNano() / 1000,
		Source:    dummy.Name(),
		Tags:      map[string]string{"tag": "tag"},
	}

	res := &DataBatch{
		Timestamp: time.Now(),
	}
	res.MetricPoints = append(res.MetricPoints, point)
	return res, nil
}

func newDummyMetricSet(name string) MetricSet {
	return MetricSet{
		MetricValues: map[string]MetricValue{},
		Labels: map[string]string{
			"name": name,
		},
	}
}

func NewDummyMetricsSource(name string, latency time.Duration) *DummyMetricsSource {
	return &DummyMetricsSource{
		latency:   latency,
		metricSet: newDummyMetricSet(name),
		name:      name,
	}
}

type DummyMetricsSourceProvider struct {
	sources           []MetricsSource
	collectionIterval time.Duration
	timeout           time.Duration
	name              string
}

func (dummy *DummyMetricsSourceProvider) GetMetricsSources() []MetricsSource {
	return dummy.sources
}

func (dummy *DummyMetricsSourceProvider) Name() string {
	return dummy.name
}

func (dummy *DummyMetricsSourceProvider) CollectionInterval() time.Duration {
	return dummy.collectionIterval
}

func (dummy *DummyMetricsSourceProvider) Timeout() time.Duration {
	return dummy.timeout
}

func NewDummyMetricsSourceProvider(name string, collectionIterval, timeout time.Duration, sources ...MetricsSource) MetricsSourceProvider {
	return &DummyMetricsSourceProvider{
		sources:           sources,
		collectionIterval: collectionIterval,
		timeout:           timeout,
		name:              name,
	}
}

type DummyDataProcessor struct {
	latency time.Duration
}

func (dummy *DummyDataProcessor) Name() string {
	return "dummy"
}

func (dummy *DummyDataProcessor) Process(data *DataBatch) (*DataBatch, error) {
	time.Sleep(dummy.latency)
	return data, nil
}

func NewDummyDataProcessor(latency time.Duration) *DummyDataProcessor {
	return &DummyDataProcessor{
		latency: latency,
	}
}

type DummyProviderHandler struct {
	count int
}

func (d *DummyProviderHandler) AddProvider(provider MetricsSourceProvider) {
	d.count += 1
}

func (d *DummyProviderHandler) DeleteProvider(name string) {
	d.count -= 1
}
