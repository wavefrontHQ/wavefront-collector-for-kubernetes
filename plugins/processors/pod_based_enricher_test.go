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
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	kube_api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var batches = []*metrics.DataBatch{
	{
		Timestamp: time.Now(),
		MetricSets: map[string]*metrics.MetricSet{
			metrics.PodContainerKey("ns1", "pod1", "c1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePodContainer,
					metrics.LabelPodName.Key:       "pod1",
					metrics.LabelNamespaceName.Key: "ns1",
					metrics.LabelContainerName.Key: "c1",
				},
				MetricValues: map[string]metrics.MetricValue{},
			},

			metrics.PodKey("ns1", "pod1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePod,
					metrics.LabelPodName.Key:       "pod1",
					metrics.LabelNamespaceName.Key: "ns1",
				},
				MetricValues: map[string]metrics.MetricValue{},
			},
		},
	},
	{
		Timestamp: time.Now(),
		MetricSets: map[string]*metrics.MetricSet{
			metrics.PodContainerKey("ns1", "pod1", "c1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePodContainer,
					metrics.LabelPodName.Key:       "pod1",
					metrics.LabelNamespaceName.Key: "ns1",
					metrics.LabelContainerName.Key: "c1",
				},
				MetricValues: map[string]metrics.MetricValue{},
			},
		},
	},
	{
		Timestamp: time.Now(),
		MetricSets: map[string]*metrics.MetricSet{
			metrics.PodKey("ns1", "pod1"): {
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePod,
					metrics.LabelPodName.Key:       "pod1",
					metrics.LabelNamespaceName.Key: "ns1",
				},
				MetricValues: map[string]metrics.MetricValue{},
			},
		},
	},
}

const otherResource = "example.com/resource1"

func TestPodEnricher(t *testing.T) {
	pod := kube_api.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "ns1",
		},
		Spec: kube_api.PodSpec{
			NodeName: "node1",
			Containers: []kube_api.Container{
				{
					Name:  "c1",
					Image: "k8s.gcr.io/pause:2.0",
					Resources: kube_api.ResourceRequirements{
						Requests: kube_api.ResourceList{
							kube_api.ResourceCPU:              *resource.NewMilliQuantity(100, resource.DecimalSI),
							kube_api.ResourceMemory:           *resource.NewQuantity(555, resource.DecimalSI),
							kube_api.ResourceEphemeralStorage: *resource.NewQuantity(1000, resource.DecimalSI),
						},
					},
				},
				{
					Name:  "nginx",
					Image: "k8s.gcr.io/pause:2.0",
					Resources: kube_api.ResourceRequirements{
						Requests: kube_api.ResourceList{
							kube_api.ResourceCPU:              *resource.NewMilliQuantity(333, resource.DecimalSI),
							kube_api.ResourceMemory:           *resource.NewQuantity(1000, resource.DecimalSI),
							kube_api.ResourceEphemeralStorage: *resource.NewQuantity(2000, resource.DecimalSI),
							otherResource:                     *resource.NewQuantity(2, resource.DecimalSI),
						},
						Limits: kube_api.ResourceList{
							kube_api.ResourceCPU:              *resource.NewMilliQuantity(2222, resource.DecimalSI),
							kube_api.ResourceMemory:           *resource.NewQuantity(3333, resource.DecimalSI),
							kube_api.ResourceEphemeralStorage: *resource.NewQuantity(5000, resource.DecimalSI),
							otherResource:                     *resource.NewQuantity(2, resource.DecimalSI),
						},
					},
				},
			},
		},
	}

	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1listers.NewPodLister(store)
	store.Add(&pod)
	labelCopier, err := util.NewLabelCopier(",", []string{}, []string{})
	assert.NoError(t, err)

	podBasedEnricher := PodBasedEnricher{
		podLister:   podLister,
		labelCopier: labelCopier,
	}

	for _, batch := range batches {
		batch, err = podBasedEnricher.Process(batch)
		assert.NoError(t, err)

		podAggregator := PodAggregator{}
		batch, err = podAggregator.Process(batch)
		assert.NoError(t, err)

		podMs, found := batch.MetricSets[metrics.PodKey("ns1", "pod1")]
		assert.True(t, found)
		checkRequests(t, podMs, 433, 1555, 3000, 2)
		checkLimits(t, podMs, 2222, 3333, 5000)

		containerMs, found := batch.MetricSets[metrics.PodContainerKey("ns1", "pod1", "c1")]
		assert.True(t, found)
		checkRequests(t, containerMs, 100, 555, 1000, -1)
		checkLimits(t, containerMs, 0, 0, 0)
	}
}

func checkRequests(t *testing.T, ms *metrics.MetricSet, cpu, mem, storage, other int64) {
	cpuVal, found := ms.MetricValues[metrics.MetricCpuRequest.Name]
	assert.True(t, found)
	assert.Equal(t, cpu, cpuVal.IntValue)

	memVal, found := ms.MetricValues[metrics.MetricMemoryRequest.Name]
	assert.True(t, found)
	assert.Equal(t, mem, memVal.IntValue)

	storageVal, found := ms.MetricValues[metrics.MetricEphemeralStorageRequest.Name]
	assert.True(t, found)
	assert.Equal(t, storage, storageVal.IntValue)

	if other > 0 {
		val, found := ms.MetricValues[otherResource+"/request"]
		assert.True(t, found)
		assert.Equal(t, other, val.IntValue)
	}
}

func checkLimits(t *testing.T, ms *metrics.MetricSet, cpu, mem int64, storage int64) {
	cpuVal, found := ms.MetricValues[metrics.MetricCpuLimit.Name]
	assert.True(t, found)
	assert.Equal(t, cpu, cpuVal.IntValue)

	memVal, found := ms.MetricValues[metrics.MetricMemoryLimit.Name]
	assert.True(t, found)
	assert.Equal(t, mem, memVal.IntValue)

	storageVal, found := ms.MetricValues[metrics.MetricEphemeralStorageLimit.Name]
	assert.True(t, found)
	assert.Equal(t, storage, storageVal.IntValue)
}
