// Based on https://github.com/kubernetes-retired/heapster/blob/master/metrics/manager/manager.go
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

// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sinks/wavefront"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources"

	log "github.com/sirupsen/logrus"
)

// FlushManager deals with data push
type FlushManager interface {
	Start()
	Stop()
}

type flushManagerImpl struct {
	processors    []metrics.DataProcessor
	sink          wavefront.WavefrontSink
	flushInterval time.Duration
	ticker        *time.Ticker
	stopChan      chan struct{}
}

// NewFlushManager crates a new PushManager
func NewFlushManager(processors []metrics.DataProcessor,
	sink wavefront.WavefrontSink, flushInterval time.Duration) (FlushManager, error) {
	manager := flushManagerImpl{
		processors:    processors,
		sink:          sink,
		flushInterval: flushInterval,
		stopChan:      make(chan struct{}),
	}

	return &manager, nil
}

func (rm *flushManagerImpl) Start() {
	rm.ticker = time.NewTicker(rm.flushInterval)
	go rm.run()
}

func (rm *flushManagerImpl) run() {
	for {
		select {
		case <-rm.ticker.C:
			go rm.push()
		case <-rm.stopChan:
			rm.ticker.Stop()
			rm.sink.Stop()
			return
		}
	}
}

func (rm *flushManagerImpl) Stop() {
	rm.stopChan <- struct{}{}
}

func (rm *flushManagerImpl) push() {
	dataBatches := sources.Manager().GetPendingMetrics()
	combinedBatch := &metrics.DataBatch{}

	for _, data := range dataBatches {
		if len(data.MetricSets) > 0 {
			// In deployment mode, the metric sets are spread across different data batches
			// as data is collected independently from each node in the cluster
			// combine all the metric sets and process them together below
			combineMetricSets(data, combinedBatch)
			continue
		}
		// Export data to sinks
		rm.sink.ExportData(data)
	}

	// process the combined metric sets
	if len(combinedBatch.MetricSets) > 0 {
		for _, p := range rm.processors {
			processedBatch, err := p.Process(combinedBatch)
			if err == nil {
				combinedBatch = processedBatch
			} else {
				log.Errorf("Error in processor: %v", err)
				return
			}
		}
		// Export to sinks
		rm.sink.ExportData(combinedBatch)
	}
}

func combineMetricSets(src, dst *metrics.DataBatch) {
	// use the most recent timestamp for the shared batch
	dst.Timestamp = src.Timestamp
	if dst.MetricSets == nil {
		dst.MetricSets = make(map[string]*metrics.MetricSet)
	}
	for k, v := range src.MetricSets {
		dst.MetricSets[k] = v
	}
}
