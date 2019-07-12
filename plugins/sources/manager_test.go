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
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
)

func init() {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	glog.SetLogger(logger)
	glog.V(0).Infof("Hi !")
}

func TestNoTimeOut(t *testing.T) {
	metricsSourceProvider := util.NewDummyMetricsSourceProvider(
		time.Minute, 100*time.Millisecond,
		util.NewDummyMetricsSource("s1", 10*time.Millisecond),
		util.NewDummyMetricsSource("s2", 10*time.Millisecond))

	manager := newEmptySourceManager()
	manager.AddProvider(metricsSourceProvider)

	time.Sleep(200 * time.Millisecond)

	dataBatchList := manager.GetPendingMetrics()

	present := make(map[string]bool)
	for _, dataBatch := range dataBatchList {
		for _, point := range dataBatch.MetricPoints {
			present[point.Metric] = true
		}
	}

	assert.True(t, present["dummy.s1"], "s1 not found - present:%v", present)
	assert.True(t, present["dummy.s1"], "s2 not found - present:%v", present)
}

func TestTimeOut(t *testing.T) {
	metricsSourceProvider := util.NewDummyMetricsSourceProvider(
		time.Minute, 75*time.Millisecond,
		util.NewDummyMetricsSource("s1", 50*time.Millisecond),
		util.NewDummyMetricsSource("s2", 100*time.Millisecond))

	manager := newEmptySourceManager()
	manager.AddProvider(metricsSourceProvider)

	time.Sleep(200 * time.Millisecond)

	dataBatchList := manager.GetPendingMetrics()

	present := make(map[string]bool)
	for _, dataBatch := range dataBatchList {
		for _, point := range dataBatch.MetricPoints {
			present[point.Metric] = true
		}
	}

	assert.True(t, present["dummy.s1"], "s1 not found - present:%v", present)
	assert.False(t, present["dummy.s2"], "s2 found - present:%v", present)
}

func TestMultipleMetrics(t *testing.T) {
	metricsSourceProvider := util.NewDummyMetricsSourceProvider(
		10*time.Millisecond, 10*time.Millisecond,
		util.NewDummyMetricsSource("s1", 0),
		util.NewDummyMetricsSource("s2", 0))

	manager := newEmptySourceManager()
	manager.AddProvider(metricsSourceProvider)

	time.Sleep(199 * time.Millisecond)

	dataBatchList := manager.GetPendingMetrics()

	counts := make(map[string]int)
	for _, dataBatch := range dataBatchList {
		for _, point := range dataBatch.MetricPoints {
			counts[point.Metric]++
		}
	}

	assert.Equal(t, 20, counts["dummy.s1"], "incorrect s1 count - counts: %vs", counts)
	assert.Equal(t, 20, counts["dummy.s2"], "incorrect s2 count - counts: %v", counts)
}
