// Based on https://github.com/kubernetes-retired/heapster/blob/master/metrics/core/types.go
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

package metrics

import (
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
)

type Type int8

const (
	Cumulative Type = iota
	Gauge
	Delta
)

func (t *Type) String() string {
	switch *t {
	case Cumulative:
		return "cumulative"
	case Gauge:
		return "gauge"
	case Delta:
		return "delta"
	}
	return ""
}

type Unit int8

const (
	Count Unit = iota
	Bytes
	Milliseconds
	Nanoseconds
	Millicores
)

func (u Unit) String() string {
	switch u {
	case Bytes:
		return "bytes"
	case Milliseconds:
		return "ms"
	case Nanoseconds:
		return "ns"
	case Millicores:
		return "millicores"
	}
	return ""
}

// ValueType is used to discriminate whether a Value is an int64 or a float64
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

// Value represents a metric value that is either a float64 or an int64
type Value struct {
	IntValue   int64
	FloatValue float64
	ValueType  ValueType
}

func (v *Value) GetValue() interface{} {
	if ValueInt64 == v.ValueType {
		return v.IntValue
	} else if ValueFloat == v.ValueType {
		return v.FloatValue
	} else {
		return nil
	}
}

// LabeledValue is a metric value that is either a float64 or an int64 and that has it's own name and values
type LabeledValue struct {
	Name   string
	Labels map[string]string
	Value
}

func (l *LabeledValue) GetValue() interface{} {
	if ValueInt64 == l.ValueType {
		return l.IntValue
	} else if ValueFloat == l.ValueType {
		return l.FloatValue
	} else {
		return nil
	}
}

// Batch contains sets of metrics tied to specific k8s resources and other more general wavefront points
type Batch struct {
	Timestamp     time.Time
	Sets          map[ResourceKey]*Set
	Points        []*wf.Point
	Distributions []*wf.Distribution
}

// Source produces metric batches
type Source interface {
	AutoDiscovered() bool
	Name() string
	Scrape() (*Batch, error)
	Cleanup()
}

// SourceProvider produces metric sources
type SourceProvider interface {
	GetMetricsSources() []Source
	Name() string
	CollectionInterval() time.Duration
	Timeout() time.Duration
}

// Sink exports metric batches
type Sink interface {
	// Export data to the external storage. The function should be synchronous/blocking and finish only
	// after the given Batch was written. This will allow sink manager to push data only to these
	// sinks that finished writing the previous data.
	Export(*Batch)
}

type Processor interface {
	Name() string
	Process(*Batch) (*Batch, error)
}

// ProviderHandler is an interface for dynamically adding and removing MetricSourceProviders
type ProviderHandler interface {
	AddProvider(provider SourceProvider)
	DeleteProvider(name string)
}

type ProviderFactory interface {
	Name() string
	Build(cfg interface{}) (SourceProvider, error)
}

type ConfigurableSourceProvider interface {
	Configure(interval, timeout time.Duration)
}

//DefaultSourceProvider handle the common providers configuration
type DefaultSourceProvider struct {
	collectionInterval time.Duration
	timeout            time.Duration
}

// CollectionInterval return the provider collection interval configuration
func (sp *DefaultSourceProvider) CollectionInterval() time.Duration {
	return sp.collectionInterval
}

// Timeout return the provider timeout configuration
func (sp *DefaultSourceProvider) Timeout() time.Duration {
	return sp.timeout
}

// Configure the 'collectionInterval' and 'timeout' values
func (sp *DefaultSourceProvider) Configure(interval, timeout time.Duration) {
	sp.collectionInterval = interval // forces default collection interval if zero
	sp.timeout = timeout
	if sp.timeout == 0 {
		sp.timeout = 30 * time.Second
	}
}
