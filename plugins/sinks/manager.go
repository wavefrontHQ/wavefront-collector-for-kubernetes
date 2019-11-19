// Based on https://github.com/kubernetes-retired/heapster/blob/master/metrics/sinks/manager.go
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

package sinks

import (
	"sync"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sinks/wavefront"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/events"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	gm "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultSinkStopTimeout = 60 * time.Second
)

var (
	sinkTimeouts gm.Counter
)

func init() {
	sinkTimeouts = gm.GetOrRegisterCounter("sink.manager.timeouts", gm.DefaultRegistry)
}

type sinkHolder struct {
	sink              wavefront.WavefrontSink
	dataBatchChannel  chan *metrics.DataBatch
	eventBatchChannel chan *events.Event
	stopChannel       chan bool
}

// Sink Manager - a special sink that distributes data to other sinks. It pushes data
// only to these sinks that completed their previous exports. Data that could not be
// pushed in the defined time is dropped and not retried.
type sinkManager struct {
	sinkHolders       []sinkHolder
	exportDataTimeout time.Duration
	stopTimeout       time.Duration
}

func NewSinkManager(sinks []wavefront.WavefrontSink, exportDataTimeout, stopTimeout time.Duration) (wavefront.WavefrontSink, error) {
	sinkHolders := []sinkHolder{}
	for _, sink := range sinks {
		sh := sinkHolder{
			sink:              sink,
			dataBatchChannel:  make(chan *metrics.DataBatch),
			eventBatchChannel: make(chan *events.Event),
			stopChannel:       make(chan bool),
		}
		sinkHolders = append(sinkHolders, sh)
		go func(sh sinkHolder) {
			for {
				select {
				case data := <-sh.dataBatchChannel:
					sh.sink.ExportData(data)
				case event := <-sh.eventBatchChannel:
					sh.sink.ExportEvent(event)
				case isStop := <-sh.stopChannel:
					log.WithField("name", sh.sink.Name()).Info("Sink stop received")
					if isStop {
						sh.sink.Stop()
						return
					}
				}
			}
		}(sh)
	}
	return &sinkManager{
		sinkHolders:       sinkHolders,
		exportDataTimeout: exportDataTimeout,
		stopTimeout:       stopTimeout,
	}, nil
}

// Guarantees that the export will complete in sinkExportDataTimeout.
func (this *sinkManager) ExportData(data *metrics.DataBatch) {
	var wg sync.WaitGroup
	for _, sh := range this.sinkHolders {
		wg.Add(1)
		go func(sh sinkHolder, wg *sync.WaitGroup) {
			defer wg.Done()
			log.WithField("name", sh.sink.Name()).Debug("Pushing data to sink")
			select {
			case sh.dataBatchChannel <- data:
				log.WithField("name", sh.sink.Name()).Info("Data push complete")
				// everything ok
			case <-time.After(this.exportDataTimeout):
				sinkTimeouts.Inc(1)
				log.WithField("name", sh.sink.Name()).Info("Data push failed")
			}
		}(sh, &wg)
	}
	// Wait for all pushes to complete or timeout.
	wg.Wait()
}

func (this *sinkManager) ExportEvent(event *events.Event) {
	var wg sync.WaitGroup
	for _, sh := range this.sinkHolders {
		wg.Add(1)
		go func(sh sinkHolder, wg *sync.WaitGroup) {
			defer wg.Done()
			log.WithField("name", sh.sink.Name()).Debug("Pushing Events to sink")
			select {
			case sh.eventBatchChannel <- event:
				log.WithField("name", sh.sink.Name()).Debug("Events push complete")
				// everything ok
			case <-time.After(this.exportDataTimeout):
				sinkTimeouts.Inc(1)
				log.WithField("name", sh.sink.Name()).Info("Events push failed")
			}
		}(sh, &wg)
	}
	// Wait for all pushes to complete or timeout.
	wg.Wait()
}

func (this *sinkManager) Name() string {
	return "Manager"
}

func (this *sinkManager) Stop() {
	for _, sh := range this.sinkHolders {
		log.Infof("Running stop for: %s", sh.sink.Name())

		go func(sh sinkHolder) {
			select {
			case sh.stopChannel <- true:
				// everything ok
				log.Infof("Stop sent to sink: %s", sh.sink.Name())

			case <-time.After(this.stopTimeout):
				log.Warningf("Failed to stop sink: %s", sh.sink.Name())
			}
			return
		}(sh)
	}
}
