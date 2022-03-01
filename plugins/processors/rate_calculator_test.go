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

package processors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

func TestRateCalculator(t *testing.T) {
	key := metrics.PodContainerKey("ns1", "pod1", "c")
	now := time.Now()

	prev := &metrics.Batch{
		Timestamp: now.Add(-time.Minute),
		Sets: map[metrics.ResourceKey]*metrics.Set{
			key: {
				CollectionStartTime: now.Add(-time.Hour),
				ScrapeTime:          now.Add(-60 * time.Second),

				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePodContainer,
				},
				Values: map[string]metrics.Value{
					metrics.MetricCpuUsage.MetricDescriptor.Name: {
						ValueType: metrics.ValueInt64,
						IntValue:  947130377781,
					},
					metrics.MetricNetworkTxErrors.MetricDescriptor.Name: {
						ValueType: metrics.ValueInt64,
						IntValue:  0,
					},
				},
			},
		},
	}

	current := &metrics.Batch{
		Timestamp: now,
		Sets: map[metrics.ResourceKey]*metrics.Set{
			key: {
				CollectionStartTime: now.Add(-time.Hour),
				ScrapeTime:          now,

				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePodContainer,
				},
				Values: map[string]metrics.Value{
					metrics.MetricCpuUsage.MetricDescriptor.Name: {
						ValueType: metrics.ValueInt64,
						IntValue:  948071062732,
					},
					metrics.MetricNetworkTxErrors.MetricDescriptor.Name: {
						ValueType: metrics.ValueInt64,
						IntValue:  120,
					},
				},
			},
		},
	}

	procesor := NewRateCalculator(metrics.RateMetricsMapping)
	procesor.Process(prev)
	procesor.Process(current)

	ms := current.Sets[key]
	cpuRate := ms.Values[metrics.MetricCpuUsageRate.Name]
	txeRate := ms.Values[metrics.MetricNetworkTxErrorsRate.Name]

	assert.InEpsilon(t, 13, cpuRate.IntValue, 2)
	assert.InEpsilon(t, 2, txeRate.FloatValue, 0.1)
}
