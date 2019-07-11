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

package manager

import (
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"github.com/golang/glog"
)

const (
	DefaultScrapeOffset   = 5 * time.Second
	DefaultMaxParallelism = 3
)

type Manager interface {
	Start()
	Stop()
}

type realManager struct {
	source                 metrics.MetricsSource
	processors             []metrics.DataProcessor
	sink                   metrics.DataSink
	resolution             time.Duration
	scrapeOffset           time.Duration
	stopChan               chan struct{}
	housekeepSemaphoreChan chan struct{}
	housekeepTimeout       time.Duration
}

func NewManager(source metrics.MetricsSource, processors []metrics.DataProcessor, sink metrics.DataSink, resolution time.Duration,
	scrapeOffset time.Duration, maxParallelism int) (Manager, error) {
	manager := realManager{
		source:                 source,
		processors:             processors,
		sink:                   sink,
		resolution:             resolution,
		scrapeOffset:           scrapeOffset,
		stopChan:               make(chan struct{}),
		housekeepSemaphoreChan: make(chan struct{}, maxParallelism),
		housekeepTimeout:       resolution / 2,
	}

	for i := 0; i < maxParallelism; i++ {
		manager.housekeepSemaphoreChan <- struct{}{}
	}

	return &manager, nil
}

func (rm *realManager) Start() {
	go rm.Housekeep()
}

func (rm *realManager) Stop() {
	rm.stopChan <- struct{}{}
}

func (rm *realManager) Handle(cfg interface{}) {
	// no-op
}

func (rm *realManager) Housekeep() {
	for {
		// Always try to get the newest metrics
		now := time.Now()
		start := now.Truncate(rm.resolution)
		end := start.Add(rm.resolution)
		timeToNextSync := end.Add(rm.scrapeOffset).Sub(now)

		select {
		case <-time.After(timeToNextSync):
			rm.housekeep(start, end)
		case <-rm.stopChan:
			rm.sink.Stop()
			return
		}
	}
}

func (rm *realManager) housekeep(start, end time.Time) {
	if !start.Before(end) {
		glog.Warningf("Wrong time provided to housekeep start:%s end: %s", start, end)
		return
	}

	select {
	case <-rm.housekeepSemaphoreChan:
		// ok, good to go

	case <-time.After(rm.housekeepTimeout):
		glog.Warningf("Spent too long waiting for housekeeping to start")
		return
	}

	go func(rm *realManager) {
		// should always give back the semaphore
		defer func() { rm.housekeepSemaphoreChan <- struct{}{} }()
		data, err := rm.source.ScrapeMetrics(start, end)

		if err != nil {
			glog.Errorf("Error in scraping metrics for %s: %v", rm.source.Name(), err)
			return
		}

		for _, p := range rm.processors {
			newData, err := process(p, data)
			if err == nil {
				data = newData
			} else {
				glog.Errorf("Error in processor: %v", err)
				return
			}
		}

		// Export data to sinks
		rm.sink.ExportData(data)
	}(rm)
}

func process(p metrics.DataProcessor, data *metrics.DataBatch) (*metrics.DataBatch, error) {
	return p.Process(data)
}
