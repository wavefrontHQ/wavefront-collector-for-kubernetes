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
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources"

	"github.com/golang/glog"
)

// PushManager deals with data push
type PushManager interface {
	Start()
	Stop()
}

type pushManagerImpl struct {
	sourceManager sources.SourceManager
	processors    []metrics.DataProcessor
	sink          metrics.DataSink
	pushInterval  time.Duration
	ticker        *time.Ticker
	stopChan      chan struct{}
}

// NewPushManager crates a new PushManager
func NewPushManager(source sources.SourceManager, processors []metrics.DataProcessor,
	sink metrics.DataSink, pushInterval time.Duration) (PushManager, error) {
	manager := pushManagerImpl{
		sourceManager: source,
		processors:    processors,
		sink:          sink,
		pushInterval:  pushInterval,
		stopChan:      make(chan struct{}),
	}

	return &manager, nil
}

func (rm *pushManagerImpl) Start() {
	rm.ticker = time.NewTicker(rm.pushInterval)
	go rm.run()
}

func (rm *pushManagerImpl) run() {
	for {
		select {
		case <-rm.ticker.C:
			go rm.push()
		case <-rm.stopChan:
			rm.ticker.Stop()
			rm.sourceManager.Stop()
			rm.sink.Stop()
			return
		}
	}
}

func (rm *pushManagerImpl) Stop() {
	rm.stopChan <- struct{}{}
}

func (rm *pushManagerImpl) push() {
	dataList := rm.sourceManager.GetPendingMetrics()
	for _, data := range dataList {
		for _, p := range rm.processors {
			newData, err := p.Process(data)
			if err == nil {
				data = newData
			} else {
				glog.Errorf("Error in processor: %v", err)
				return
			}
		}
		// Export data to sinks
		rm.sink.ExportData(data)
	}
}
