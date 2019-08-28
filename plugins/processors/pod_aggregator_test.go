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

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

func TestPodAggregator(t *testing.T) {
	batch := metrics.DataBatch{
		Timestamp: time.Now(),
		MetricSets: map[string]*metrics.MetricSet{
			metrics.PodContainerKey("ns1", "pod1", "c1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePodContainer,
					metrics.LabelPodName.Key:       "pod1",
					metrics.LabelNamespaceName.Key: "ns1",
				},
				MetricValues: map[string]metrics.MetricValue{
					"m1": {
						ValueType:  metrics.ValueInt64,
						MetricType: metrics.MetricGauge,
						IntValue:   10,
					},
					"m2": {
						ValueType:  metrics.ValueInt64,
						MetricType: metrics.MetricGauge,
						IntValue:   222,
					},
				},
			},

			metrics.PodContainerKey("ns1", "pod1", "c2"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePodContainer,
					metrics.LabelPodName.Key:       "pod1",
					metrics.LabelNamespaceName.Key: "ns1",
				},
				MetricValues: map[string]metrics.MetricValue{
					"m1": {
						ValueType:  metrics.ValueInt64,
						MetricType: metrics.MetricGauge,
						IntValue:   100,
					},
					"m3": {
						ValueType:  metrics.ValueInt64,
						MetricType: metrics.MetricGauge,
						IntValue:   30,
					},
				},
			},

			metrics.PodKey("ns1", "pod2"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePod,
					metrics.LabelPodName.Key:       "pod2",
					metrics.LabelNamespaceName.Key: "ns1",
				},
				MetricValues: map[string]metrics.MetricValue{
					"m1": {
						ValueType:  metrics.ValueInt64,
						MetricType: metrics.MetricGauge,
						IntValue:   100,
					},
				},
			},

			metrics.PodContainerKey("ns1", "pod2", "c1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePodContainer,
					metrics.LabelPodName.Key:       "pod2",
					metrics.LabelNamespaceName.Key: "ns1",
				},
				MetricValues: map[string]metrics.MetricValue{
					"m1": {
						ValueType:  metrics.ValueInt64,
						MetricType: metrics.MetricGauge,
						IntValue:   10,
					},
					"m2": {
						ValueType:  metrics.ValueInt64,
						MetricType: metrics.MetricGauge,
						IntValue:   20,
					},
				},
			},
		},
	}
	processor := PodAggregator{}
	result, err := processor.Process(&batch)
	assert.NoError(t, err)
	pod, found := result.MetricSets[metrics.PodKey("ns1", "pod1")]
	assert.True(t, found)

	m1, found := pod.MetricValues["m1"]
	assert.True(t, found)
	assert.Equal(t, int64(110), m1.IntValue)

	m2, found := pod.MetricValues["m2"]
	assert.True(t, found)
	assert.Equal(t, int64(222), m2.IntValue)

	m3, found := pod.MetricValues["m3"]
	assert.True(t, found)
	assert.Equal(t, int64(30), m3.IntValue)

	labelPodName, found := pod.Labels[metrics.LabelPodName.Key]
	assert.True(t, found)
	assert.Equal(t, "pod1", labelPodName)

	labelNsName, found := pod.Labels[metrics.LabelNamespaceName.Key]
	assert.True(t, found)
	assert.Equal(t, "ns1", labelNsName)

	pod, found = result.MetricSets[metrics.PodKey("ns1", "pod2")]
	assert.True(t, found)

	m1, found = pod.MetricValues["m1"]
	assert.True(t, found)
	assert.Equal(t, int64(100), m1.IntValue)

	m2, found = pod.MetricValues["m2"]
	assert.True(t, found)
	assert.Equal(t, int64(20), m2.IntValue)

}
