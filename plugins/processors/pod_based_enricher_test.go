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

const otherResource = "example.com/resource1"

type enricherTestContext struct {
	pod                *kube_api.Pod
	batch              *metrics.DataBatch
	collectionInterval time.Duration
}

func TestPodEnricher(t *testing.T) {
	tc := setup()
	podBasedEnricher := createEnricher(t, tc)

	var batches = []*metrics.DataBatch{
		createContainerBatch(),
		createContainerBatch(),
		createPodBatch(),
	}

	var err error
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

func TestStatusRunning(t *testing.T) {
	tc := setup()
	tc.pod.Status = kube_api.PodStatus{
		ContainerStatuses: []kube_api.ContainerStatus{
			{
				Name:  "c1",
				State: createCrashState(time.Now().Add(-10*time.Minute), time.Now().Add(-5*time.Minute)),
			},
		},
	}

	podBasedEnricher := createEnricher(t, tc)

	batch, err := podBasedEnricher.Process(tc.batch)
	assert.NoError(t, err)

	containerMs, found := batch.MetricSets[metrics.PodContainerKey("ns1", "pod1", "c1")]
	assert.True(t, found)

	expectedStatus := metrics.LabeledMetric{
		Name: "status",
		Labels: map[string]string{
			"status": "terminated",
			"reason": "bad juju",
		},
		MetricValue: metrics.MetricValue{
			IntValue:   3,
			MetricType: metrics.MetricGauge,
		},
	}
	assert.Equal(t, expectedStatus, containerMs.LabeledMetrics[0])
}

func TestStatusMissedTermination(t *testing.T) {
	tc := setup()

	now := time.Now()
	firstStart := now.Add(-10 * time.Minute)
	crashTime := now.Add(-30 * time.Second)
	latestStart := now.Add(-5 * time.Second)

	missedCollectionTime := now

	tc.pod.Status = kube_api.PodStatus{
		ContainerStatuses: []kube_api.ContainerStatus{
			{
				Name:                 "c1",
				State:                createGoodState(latestStart),
				LastTerminationState: createCrashState(firstStart, crashTime),
			},
		},
	}

	podBasedEnricher := createEnricher(t, tc)

	tc.batch.Timestamp = missedCollectionTime
	expectedStatus := metrics.LabeledMetric{
		Name: "status",
		Labels: map[string]string{
			"status": "terminated",
			"reason": "bad juju",
		},
		MetricValue: metrics.MetricValue{
			IntValue:   3,
			MetricType: metrics.MetricGauge,
		},
	}
	assert.Equal(t, expectedStatus, processBatch(t, podBasedEnricher, tc.batch))
}

func TestStatusPassedTermination(t *testing.T) {
	tc := setup()

	now := time.Now()
	firstStart := now.Add(-10 * time.Minute)
	crashTime := now.Add(-30 * time.Second)
	latestStart := now.Add(-5 * time.Second)

	followingCollectionTime := now.Add(tc.collectionInterval)

	tc.pod.Status = kube_api.PodStatus{
		ContainerStatuses: []kube_api.ContainerStatus{
			{
				Name:                 "c1",
				State:                createGoodState(latestStart),
				LastTerminationState: createCrashState(firstStart, crashTime),
			},
		},
	}

	podBasedEnricher := createEnricher(t, tc)

	expectedStatus := metrics.LabeledMetric{
		Name: "status",
		Labels: map[string]string{
			"status": "running",
		},
		MetricValue: metrics.MetricValue{
			IntValue:   1,
			MetricType: metrics.MetricGauge,
		},
	}
	batch2 := createContainerBatch()
	batch2.Timestamp = followingCollectionTime
	assert.Equal(t, expectedStatus, processBatch(t, podBasedEnricher, batch2))
}

func processBatch(t assert.TestingT, podBasedEnricher *PodBasedEnricher, batch *metrics.DataBatch) metrics.LabeledMetric {
	var err error
	batch, err = podBasedEnricher.Process(batch)
	assert.NoError(t, err)

	containerMs, found := batch.MetricSets[metrics.PodContainerKey("ns1", "pod1", "c1")]
	assert.True(t, found)
	return containerMs.LabeledMetrics[0]
}

func createEnricher(t *testing.T, tc *enricherTestContext) *PodBasedEnricher {
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1listers.NewPodLister(store)
	err := store.Add(tc.pod)
	assert.NoError(t, err)

	labelCopier, err := util.NewLabelCopier(",", []string{}, []string{})
	assert.NoError(t, err)

	return NewPodBasedEnricher(podLister, labelCopier, tc.collectionInterval)
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

func setup() *enricherTestContext {
	return &enricherTestContext{
		collectionInterval: time.Minute,
		batch:              createContainerBatch(),
		pod: &kube_api.Pod{
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
		},
	}
}

func createContainerBatch() *metrics.DataBatch {
	return &metrics.DataBatch{
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
	}
}

func createPodBatch() *metrics.DataBatch {
	return &metrics.DataBatch{
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
	}
}

func createCrashState(startTime time.Time, crashTime time.Time) kube_api.ContainerState {
	return kube_api.ContainerState{
		Terminated: &kube_api.ContainerStateTerminated{
			Reason:  "bad juju",
			Message: "broken",
			StartedAt: metav1.Time{
				Time: startTime,
			},
			FinishedAt: metav1.Time{
				Time: crashTime,
			},
			ContainerID: "",
		},
	}
}

func createGoodState(timestamp time.Time) kube_api.ContainerState {
	return kube_api.ContainerState{
		Running: &kube_api.ContainerStateRunning{
			StartedAt: metav1.Time{
				Time: timestamp,
			},
		},
	}
}
