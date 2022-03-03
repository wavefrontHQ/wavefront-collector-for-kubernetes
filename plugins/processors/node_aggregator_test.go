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

	corev1 "k8s.io/api/core/v1"
)

func TestNodeAggregate(t *testing.T) {
	batch := metrics.Batch{
		Timestamp: time.Now(),
		Sets: map[metrics.ResourceKey]*metrics.Set{
			metrics.PodKey("ns1", "pod1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePod,
					metrics.LabelNamespaceName.Key: "ns1",
					metrics.LabelNodename.Key:      "h1",
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
				},
				LabeledValues: []metrics.LabeledValue{
					{
						Name:   metrics.MetricPodPhase.Name,
						Labels: map[string]string{"phase": string(corev1.PodSucceeded)},
						Value: metrics.Value{
							ValueType: metrics.ValueInt64,
							IntValue:  convertPhase(corev1.PodSucceeded),
						},
					},
				},
			},

			metrics.PodContainerKey("ns1", "pod1", "container1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePodContainer,
					metrics.LabelNamespaceName.Key: "ns1",
					metrics.LabelNodename.Key:      "h1",
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
				},
				LabeledValues: []metrics.LabeledValue{
					{
						Name:   metrics.MetricContainerStatus.Name,
						Labels: map[string]string{"state": "terminated"},
						Value: metrics.Value{
							ValueType: metrics.ValueInt64,
							IntValue:  3,
						},
					},
				},
			},

			metrics.PodKey("ns1", "pod2"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePod,
					metrics.LabelNamespaceName.Key: "ns1",
					metrics.LabelNodename.Key:      "h1",
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
				},
			},

			metrics.PodContainerKey("ns1", "pod2", "container2"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePodContainer,
					metrics.LabelNamespaceName.Key: "ns1",
					metrics.LabelNodename.Key:      "h1",
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
				},
			},

			metrics.NodeKey("h1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypeNode,
					metrics.LabelNodename.Key:      "h1",
				},
				Values: map[string]metrics.Value{},
			},
		},
	}
	processor := NodeAggregator{
		MetricsToAggregate: []string{"m1", "m3"},
	}
	result, err := processor.Process(&batch)
	assert.NoError(t, err)
	node, found := result.Sets[metrics.NodeKey("h1")]
	assert.True(t, found)

	m1, found := node.Values["m1"]
	assert.True(t, found)
	assert.Equal(t, int64(100), m1.IntValue)

	m3, found := node.Values["m3"]
	assert.True(t, found)
	assert.Equal(t, int64(30), m3.IntValue)

	podCount, found := node.Values[metrics.MetricPodCount.Name]
	assert.True(t, found)
	assert.Equal(t, int64(1), podCount.IntValue)

	podContainerCount, found := node.Values[metrics.MetricPodContainerCount.Name]
	assert.True(t, found)
	assert.Equal(t, int64(1), podContainerCount.IntValue)
}
