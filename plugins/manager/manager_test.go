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

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources"
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
