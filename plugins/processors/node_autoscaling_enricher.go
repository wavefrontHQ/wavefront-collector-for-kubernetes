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
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	kube_api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	kube_client "k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type NodeAutoscalingEnricher struct {
	nodeLister  v1listers.NodeLister
	reflector   *cache.Reflector
	labelCopier *util.LabelCopier
}

func (nae *NodeAutoscalingEnricher) Name() string {
	return "node_autoscaling_enricher"
}

func (nae *NodeAutoscalingEnricher) Process(batch *metrics.DataBatch) (*metrics.DataBatch, error) {
	nodes, err := nae.nodeLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if metricSet, found := batch.MetricSets[metrics.NodeKey(node.Name)]; found {
			nae.labelCopier.Copy(node.Labels, metricSet.Labels)
			capacityCpu, _ := node.Status.Capacity[kube_api.ResourceCPU]
			capacityMem, _ := node.Status.Capacity[kube_api.ResourceMemory]
			capacityEphemeralStorage, storageExist := node.Status.Capacity[kube_api.ResourceEphemeralStorage]
			allocatableCpu, _ := node.Status.Allocatable[kube_api.ResourceCPU]
			allocatableMem, _ := node.Status.Allocatable[kube_api.ResourceMemory]
			allocatableEphemeralStorage, allocatableStorageExist := node.Status.Allocatable[kube_api.ResourceEphemeralStorage]

			cpuRequested := getInt(metricSet, &metrics.MetricCpuRequest)
			cpuUsed := getInt(metricSet, &metrics.MetricCpuUsageRate)
			memRequested := getInt(metricSet, &metrics.MetricMemoryRequest)
			memWorkingSet := getInt(metricSet, &metrics.MetricMemoryWorkingSet)
			epheRequested := getInt(metricSet, &metrics.MetricEphemeralStorageRequest)
			epheUsed := getInt(metricSet, &metrics.MetricEphemeralStorageUsage)

			if allocatableCpu.MilliValue() != 0 {
				setFloat(metricSet, &metrics.MetricNodeCpuUtilization, float64(cpuUsed)/float64(allocatableCpu.MilliValue()))
				setFloat(metricSet, &metrics.MetricNodeCpuReservation, float64(cpuRequested)/float64(allocatableCpu.MilliValue()))
			}
			setFloat(metricSet, &metrics.MetricNodeCpuCapacity, float64(capacityCpu.MilliValue()))
			setFloat(metricSet, &metrics.MetricNodeCpuAllocatable, float64(allocatableCpu.MilliValue()))

			if allocatableMem.Value() != 0 {
				setFloat(metricSet, &metrics.MetricNodeMemoryUtilization, float64(memWorkingSet)/float64(allocatableMem.Value()))
				setFloat(metricSet, &metrics.MetricNodeMemoryReservation, float64(memRequested)/float64(allocatableMem.Value()))
			}
			setFloat(metricSet, &metrics.MetricNodeMemoryCapacity, float64(capacityMem.Value()))
			setFloat(metricSet, &metrics.MetricNodeMemoryAllocatable, float64(allocatableMem.Value()))

			if storageExist && allocatableStorageExist {
				setFloat(metricSet, &metrics.MetricNodeEphemeralStorageCapacity, float64(capacityEphemeralStorage.Value()))
				setFloat(metricSet, &metrics.MetricNodeEphemeralStorageAllocatable, float64(allocatableEphemeralStorage.Value()))
				if allocatableEphemeralStorage.Value() != 0 {
					setFloat(metricSet, &metrics.MetricNodeEphemeralStorageUtilization, float64(epheUsed)/float64(allocatableEphemeralStorage.Value()))
					setFloat(metricSet, &metrics.MetricNodeEphemeralStorageReservation, float64(epheRequested)/float64(allocatableEphemeralStorage.Value()))
				}
			}
		}
	}
	return batch, nil
}

func getInt(metricSet *metrics.MetricSet, metric *metrics.Metric) int64 {
	if value, found := metricSet.MetricValues[metric.MetricDescriptor.Name]; found {
		return value.IntValue
	}
	return 0
}

func setFloat(metricSet *metrics.MetricSet, metric *metrics.Metric, value float64) {
	metricSet.MetricValues[metric.MetricDescriptor.Name] = metrics.MetricValue{
		ValueType:  metrics.ValueFloat,
		FloatValue: value,
	}
}

func NewNodeAutoscalingEnricher(kubeClient *kube_client.Clientset, labelCopier *util.LabelCopier) (*NodeAutoscalingEnricher, error) {
	// watch nodes
	nodeLister, reflector, _ := util.GetNodeLister(kubeClient)

	return &NodeAutoscalingEnricher{
		nodeLister:  nodeLister,
		reflector:   reflector,
		labelCopier: labelCopier,
	}, nil
}
