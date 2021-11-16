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

func TestNodeAggregate(t *testing.T) {
	batch := metrics.DataBatch{
		Timestamp: time.Now(),
		MetricSets: map[string]*metrics.MetricSet{
			metrics.PodKey("ns1", "pod1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePod,
					metrics.LabelNamespaceName.Key: "ns1",
					metrics.LabelNodename.Key:      "h1",
				},
				MetricValues: map[string]metrics.MetricValue{
					"m1": {
						ValueType: metrics.ValueInt64,
						IntValue:  10,
					},
					"m2": {
						ValueType: metrics.ValueInt64,
						IntValue:  222,
					},
				},
			},

			metrics.PodKey("ns1", "pod2"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePod,
					metrics.LabelNamespaceName.Key: "ns1",
					metrics.LabelNodename.Key:      "h1",
				},
				MetricValues: map[string]metrics.MetricValue{
					"m1": {
						ValueType: metrics.ValueInt64,
						IntValue:  100,
					},
					"m3": {
						ValueType: metrics.ValueInt64,
						IntValue:  30,
					},
				},
			},

			metrics.NodeKey("h1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypeNode,
					metrics.LabelNodename.Key:      "h1",
				},
				MetricValues: map[string]metrics.MetricValue{},
			},
		},
	}
	processor := NodeAggregator{
		MetricsToAggregate: []string{"m1", "m3"},
	}
	result, err := processor.Process(&batch)
	assert.NoError(t, err)
	node, found := result.MetricSets[metrics.NodeKey("h1")]
	assert.True(t, found)

	m1, found := node.MetricValues["m1"]
	assert.True(t, found)
	assert.Equal(t, int64(110), m1.IntValue)

	m3, found := node.MetricValues["m3"]
	assert.True(t, found)
	assert.Equal(t, int64(30), m3.IntValue)
}
