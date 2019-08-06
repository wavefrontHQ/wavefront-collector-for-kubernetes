// Based on https://github.com/kubernetes-retired/heapster/blob/master/metrics/core/types.go
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

// enhanced to support Wavefront data format MetricPoints

package metrics

import (
	"time"
)

type MetricType int8

const (
	MetricCumulative MetricType = iota
	MetricGauge
	MetricDelta
)

func (self *MetricType) String() string {
	switch *self {
	case MetricCumulative:
		return "cumulative"
	case MetricGauge:
		return "gauge"
	case MetricDelta:
		return "delta"
	}
	return ""
}

type ValueType int8

const (
	ValueInt64 ValueType = iota
	ValueFloat
)

func (valueType *ValueType) String() string {
	switch *valueType {
	case ValueInt64:
		return "int64"
	case ValueFloat:
		return "double"
	}
	return ""
}

type UnitsType int8

const (
	// A counter metric.
	UnitsCount UnitsType = iota
	// A metric in bytes.
	UnitsBytes
	// A metric in milliseconds.
	UnitsMilliseconds
	// A metric in nanoseconds.
	UnitsNanoseconds
	// A metric in millicores.
	UnitsMillicores
)

func (unitsType *UnitsType) String() string {
	switch *unitsType {
	case UnitsBytes:
		return "bytes"
	case UnitsMilliseconds:
		return "ms"
	case UnitsNanoseconds:
		return "ns"
	case UnitsMillicores:
		return "millicores"
	}
	return ""
}

type MetricValue struct {
	IntValue   int64
	FloatValue float64
	MetricType MetricType
	ValueType  ValueType
}

func (metricValue *MetricValue) GetValue() interface{} {
	if ValueInt64 == metricValue.ValueType {
		return metricValue.IntValue
	} else if ValueFloat == metricValue.ValueType {
		return metricValue.FloatValue
	} else {
		return nil
	}
}

type LabeledMetric struct {
	Name   string
	Labels map[string]string
	MetricValue
}

func (labeledMetric *LabeledMetric) GetValue() interface{} {
	if ValueInt64 == labeledMetric.ValueType {
		return labeledMetric.IntValue
	} else if ValueFloat == labeledMetric.ValueType {
		return labeledMetric.FloatValue
	} else {
		return nil
	}
}

type MetricSet struct {
	// CollectionStartTime is a time since when the metrics are collected for this entity.
	// It is affected by events like entity (e.g. pod) creation, entity restart (e.g. for container),
	// Kubelet restart.
	CollectionStartTime time.Time
	// EntityCreateTime is a time of entity creation and persists through entity restarts and
	// Kubelet restarts.
	EntityCreateTime time.Time
	ScrapeTime       time.Time
	MetricValues     map[string]MetricValue
	Labels           map[string]string
	LabeledMetrics   []LabeledMetric
}

type DataBatch struct {
	Timestamp time.Time
	// Should use key functions from ms_keys.go
	MetricSets   map[string]*MetricSet
	MetricPoints []*MetricPoint
}

// A place from where the metrics should be scraped.
type MetricsSource interface {
	Name() string
	ScrapeMetrics() (*DataBatch, error)
}

// Provider of list of sources to be scraped.
type MetricsSourceProvider interface {
	GetMetricsSources() []MetricsSource
	Name() string
	CollectionInterval() time.Duration
	Timeout() time.Duration
}

type DataSink interface {
	Name() string

	// Exports data to the external storage. The function should be synchronous/blocking and finish only
	// after the given DataBatch was written. This will allow sink manager to push data only to these
	// sinks that finished writing the previous data.
	ExportData(*DataBatch)
	Stop()
}

type DataProcessor interface {
	Name() string
	Process(*DataBatch) (*DataBatch, error)
}

// Represents a single point in Wavefront metric format.
type MetricPoint struct {
	Metric    string
	Value     float64
	Timestamp int64
	Source    string
	Tags      map[string]string
	StrTags   string
}

// ProviderHandler is an interface for dynamically adding and removing MetricSourceProviders
type ProviderHandler interface {
	AddProvider(provider MetricsSourceProvider)
	DeleteProvider(name string)
}

type ProviderFactory interface {
	Name() string
	Build(cfg interface{}) (MetricsSourceProvider, error)
}

type ConfigurabeMetricsSourceProvider interface {
	Configure(interval, timeout time.Duration)
}

//DefaultMetricsSourceProvider handle the common providers configuration
type DefaultMetricsSourceProvider struct {
	collectionInterval time.Duration
	timeout            time.Duration
}

// CollectionInterval return the provider collection interval configuration
func (dp *DefaultMetricsSourceProvider) CollectionInterval() time.Duration {
	return dp.collectionInterval
}

// Timeout return the provider timeout configuration
func (dp *DefaultMetricsSourceProvider) Timeout() time.Duration {
	return dp.timeout
}

// Configure the 'collectionInterval' and 'timeout' values
func (dp *DefaultMetricsSourceProvider) Configure(interval, timeout time.Duration) {
	dp.collectionInterval = interval // forces default collection interval if zero
	dp.timeout = timeout
	if dp.timeout == 0 {
		dp.timeout = 10 * time.Second
	}
}
