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

func TestClusterAggregate(t *testing.T) {
	batch := metrics.Batch{
		Timestamp: time.Now(),
		Sets: map[metrics.ResourceKey]*metrics.Set{
			metrics.NamespaceKey("ns1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypeNamespace,
					metrics.LabelNamespaceName.Key: "ns1",
				},
				Values: map[string]metrics.Value{
					"m1": {
						ValueType: metrics.ValueInt64,
						IntValue:  10,
					},
					"m2": {
						ValueType: metrics.ValueInt64,
						IntValue:  222,
					},
                    metrics.MetricPodCount.Name: {
                        ValueType: metrics.ValueInt64,
                        IntValue: 1,
                    },
                    metrics.MetricPodContainerCount.Name: {
                        ValueType: metrics.ValueInt64,
                        IntValue: 1,
                    },
				},
			},

			metrics.NamespaceKey("ns2"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypeNamespace,
					metrics.LabelNamespaceName.Key: "ns1",
				},
				Values: map[string]metrics.Value{
					"m1": {
						ValueType: metrics.ValueInt64,
						IntValue:  100,
					},
					"m3": {
						ValueType: metrics.ValueInt64,
						IntValue:  30,
					},
                    metrics.MetricPodCount.Name: {
                        ValueType: metrics.ValueInt64,
                        IntValue: 2,
                    },
                    metrics.MetricPodContainerCount.Name: {
                        ValueType: metrics.ValueInt64,
                        IntValue: 2,
                    },
				},
			},
		},
	}
	processor := ClusterAggregator{
		MetricsToAggregate: []string{"m1", "m3"},
	}
	result, err := processor.Process(&batch)
	assert.NoError(t, err)
	cluster, found := result.Sets[metrics.ClusterKey()]
	assert.True(t, found)

	m1, found := cluster.Values["m1"]
	assert.True(t, found)
	assert.Equal(t, int64(110), m1.IntValue)

	m3, found := cluster.Values["m3"]
	assert.True(t, found)
	assert.Equal(t, int64(30), m3.IntValue)

    podCount, found := cluster.Values[metrics.MetricPodCount.Name]
    assert.True(t, found)
    assert.Equal(t, int64(3), podCount.IntValue)

    podContainerCount, found := cluster.Values[metrics.MetricPodContainerCount.Name]
    assert.True(t, found)
    assert.Equal(t, int64(3), podContainerCount.IntValue)
}
