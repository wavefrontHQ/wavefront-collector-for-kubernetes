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

package util

import (
	"strings"
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

type DummySink struct {
	name        string
	mutex       sync.Mutex
	exportCount int
	stopped     bool
	latency     time.Duration
}

func (this *DummySink) Name() string {
	return this.name
}
func (this *DummySink) ExportData(*DataBatch) {
	this.mutex.Lock()
	this.exportCount++
	this.mutex.Unlock()

	time.Sleep(this.latency)
}

func (this *DummySink) Stop() {
	this.mutex.Lock()
	this.stopped = true
	this.mutex.Unlock()

	time.Sleep(this.latency)
}

func (this *DummySink) IsStopped() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.stopped
}

func (this *DummySink) GetExportCount() int {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.exportCount
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

func (this *DummyMetricsSource) Name() string {
	return this.name
}

func (this *DummyMetricsSource) ScrapeMetrics() (*DataBatch, error) {
	time.Sleep(this.latency)

	point := &metrics.MetricPoint{
		Metric:    strings.Replace(this.Name(), " ", ".", -1),
		Value:     1,
		Timestamp: time.Now().UnixNano() / 1000,
		Source:    this.Name(),
		Tags:      map[string]string{"tag": "tag"},
	}

	res := &DataBatch{
		Timestamp:  time.Now(),
		MetricSets: map[string]*MetricSet{},
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

func (this *DummyMetricsSourceProvider) GetMetricsSources() []MetricsSource {
	return this.sources
}

func (this *DummyMetricsSourceProvider) Name() string {
	return this.name
}

func (this *DummyMetricsSourceProvider) CollectionInterval() time.Duration {
	return this.collectionIterval
}

func (this *DummyMetricsSourceProvider) Timeout() time.Duration {
	return this.timeout
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

func (this *DummyDataProcessor) Name() string {
	return "dummy"
}

func (this *DummyDataProcessor) Process(data *DataBatch) (*DataBatch, error) {
	time.Sleep(this.latency)
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
