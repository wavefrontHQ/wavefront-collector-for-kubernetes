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
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	kube_api "k8s.io/api/core/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
)

type PodBasedEnricher struct {
	podLister          v1listers.PodLister
	labelCopier        *util.LabelCopier
	collectionInterval time.Duration
}

func (pbe *PodBasedEnricher) Name() string {
	return "pod_based_enricher"
}

func (pbe *PodBasedEnricher) Process(batch *metrics.Batch) (*metrics.Batch, error) {
	newMs := make(map[metrics.ResourceKey]*metrics.Set, len(batch.Sets))
	for k, v := range batch.Sets {
		switch v.Labels[metrics.LabelMetricSetType.Key] {
		case metrics.MetricSetTypePod:
			namespace := v.Labels[metrics.LabelNamespaceName.Key]
			podName := v.Labels[metrics.LabelPodName.Key]
			pod, err := pbe.getPod(namespace, podName)
			if err != nil {
				delete(batch.Sets, k)
				log.Debugf("Failed to get pod %s from cache: %v", metrics.PodKey(namespace, podName), err)
				continue
			}
			pbe.addPodInfo(v, pod, batch, newMs)
		case metrics.MetricSetTypePodContainer:
			namespace := v.Labels[metrics.LabelNamespaceName.Key]
			podName := v.Labels[metrics.LabelPodName.Key]
			pod, err := pbe.getPod(namespace, podName)
			if err != nil {
				delete(batch.Sets, k)
				log.Debugf("Failed to get pod %s from cache: %v", metrics.PodKey(namespace, podName), err)
				continue
			}
			pbe.addContainerInfo(k, v, pod, batch, newMs)
		}
	}
	for k, v := range newMs {
		batch.Sets[k] = v
	}
	return batch, nil
}

func (pbe *PodBasedEnricher) getPod(namespace, name string) (*kube_api.Pod, error) {
	pod, err := pbe.podLister.Pods(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	if pod == nil {
		return nil, fmt.Errorf("cannot find pod definition")
	}
	return pod, nil
}

func (pbe *PodBasedEnricher) addContainerInfo(key metrics.ResourceKey, containerMs *metrics.Set, pod *kube_api.Pod, batch *metrics.Batch, newMs map[metrics.ResourceKey]*metrics.Set) {
	for _, container := range pod.Spec.Containers {
		if key == metrics.PodContainerKey(pod.Namespace, pod.Name, container.Name) {
			updateContainerResourcesAndLimits(containerMs, container)
			if _, ok := containerMs.Labels[metrics.LabelContainerBaseImage.Key]; !ok {
				containerMs.Labels[metrics.LabelContainerBaseImage.Key] = container.Image
			}
			break
		}
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if key == metrics.PodContainerKey(pod.Namespace, pod.Name, containerStatus.Name) {
			containerMs.Values[metrics.MetricRestartCount.Name] = intValue(int64(containerStatus.RestartCount))
			if !pod.Status.StartTime.IsZero() {
				containerMs.EntityCreateTime = pod.Status.StartTime.Time
			}
			pbe.addContainerStatus(batch.Timestamp, containerMs, &metrics.MetricContainerStatus, containerStatus)
			break
		}
	}

	containerMs.Labels[metrics.LabelPodId.Key] = string(pod.UID)
	pbe.labelCopier.Copy(pod.Labels, containerMs.Labels)

	namespace := containerMs.Labels[metrics.LabelNamespaceName.Key]
	podName := containerMs.Labels[metrics.LabelPodName.Key]

	podKey := metrics.PodKey(namespace, podName)
	_, oldfound := batch.Sets[podKey]
	if !oldfound {
		_, newfound := newMs[podKey]
		if !newfound {
			log.Debugf("Pod %s not found, creating a stub", podKey)
			podMs := &metrics.Set{
				Values: make(map[string]metrics.Value),
				Labels: map[string]string{
					metrics.LabelMetricSetType.Key: metrics.MetricSetTypePod,
					metrics.LabelNamespaceName.Key: namespace,
					metrics.LabelPodName.Key:       podName,
					metrics.LabelNodename.Key:      containerMs.Labels[metrics.LabelNodename.Key],
					metrics.LabelHostname.Key:      containerMs.Labels[metrics.LabelHostname.Key],
					metrics.LabelHostID.Key:        containerMs.Labels[metrics.LabelHostID.Key],
				},
			}
			if !pod.Status.StartTime.IsZero() {
				podMs.EntityCreateTime = pod.Status.StartTime.Time
			}
			newMs[podKey] = podMs
			pbe.addPodInfo(podMs, pod, batch, newMs)
		}
	}
}

func (pbe *PodBasedEnricher) addPodInfo(podMs *metrics.Set, pod *kube_api.Pod, batch *metrics.Batch, newMs map[metrics.ResourceKey]*metrics.Set) {
	// Add UID and create time to pod
	podMs.Labels[metrics.LabelPodId.Key] = string(pod.UID)
	if !pod.Status.StartTime.IsZero() {
		podMs.EntityCreateTime = pod.Status.StartTime.Time
	}
	pbe.labelCopier.Copy(pod.Labels, podMs.Labels)

	// Add pod phase
	addLabeledIntMetric(podMs, &metrics.MetricPodPhase, map[string]string{"phase": string(pod.Status.Phase)}, convertPhase(pod.Status.Phase))

	// Add cpu/mem requests and limits to containers
	for _, container := range pod.Spec.Containers {
		containerKey := metrics.PodContainerKey(pod.Namespace, pod.Name, container.Name)
		if _, found := batch.Sets[containerKey]; found {
			continue
		}
		if _, found := newMs[containerKey]; found {
			continue
		}
		log.Debugf("Container %s not found, creating a stub", containerKey)
		containerMs := &metrics.Set{
			Values: make(map[string]metrics.Value),
			Labels: map[string]string{
				metrics.LabelMetricSetType.Key:      metrics.MetricSetTypePodContainer,
				metrics.LabelNamespaceName.Key:      pod.Namespace,
				metrics.LabelPodName.Key:            pod.Name,
				metrics.LabelContainerName.Key:      container.Name,
				metrics.LabelContainerBaseImage.Key: container.Image,
				metrics.LabelPodId.Key:              string(pod.UID),
				metrics.LabelNodename.Key:           podMs.Labels[metrics.LabelNodename.Key],
				metrics.LabelHostname.Key:           podMs.Labels[metrics.LabelHostname.Key],
				metrics.LabelHostID.Key:             podMs.Labels[metrics.LabelHostID.Key],
			},
			EntityCreateTime: podMs.CollectionStartTime,
		}
		pbe.labelCopier.Copy(pod.Labels, containerMs.Labels)
		updateContainerResourcesAndLimits(containerMs, container)
		newMs[containerKey] = containerMs
	}
	pbe.updateContainerStatus(newMs, pod, pod.Status.ContainerStatuses, batch.Timestamp)
}

func updateContainerResourcesAndLimits(metricSet *metrics.Set, container kube_api.Container) {
	requests := container.Resources.Requests

	for key, val := range container.Resources.Requests {
		metric, found := metrics.ResourceRequestMetrics[key]
		// Inserts a metric to metrics.ResourceRequestMetrics if there is no
		// existing one for the given resource. The name of this metric is
		// ResourceName/request where ResourceName is the name of the resource
		// requested in container resource requests.
		if !found {
			metric = metrics.Metric{
				MetricDescriptor: metrics.MetricDescriptor{
					Name:        string(key) + "/request",
					Description: string(key) + " resource request. This metric is Kubernetes specific.",
					Type:        metrics.Gauge,
					ValueType:   metrics.ValueInt64,
					Units:       metrics.Count,
				},
			}
			metrics.ResourceRequestMetrics[key] = metric
		}
		if key == kube_api.ResourceCPU {
			metricSet.Values[metric.Name] = intValue(val.MilliValue())
		} else {
			metricSet.Values[metric.Name] = intValue(val.Value())
		}
	}

	// For primary resources like cpu and memory, explicitly sets their request resource
	// metric to zero if they are not requested.
	if _, found := requests[kube_api.ResourceCPU]; !found {
		metricSet.Values[metrics.MetricCpuRequest.Name] = intValue(0)
	}
	if _, found := requests[kube_api.ResourceMemory]; !found {
		metricSet.Values[metrics.MetricMemoryRequest.Name] = intValue(0)
	}
	if _, found := requests[kube_api.ResourceEphemeralStorage]; !found {
		metricSet.Values[metrics.MetricEphemeralStorageRequest.Name] = intValue(0)
	}

	limits := container.Resources.Limits
	if val, found := limits[kube_api.ResourceCPU]; found {
		metricSet.Values[metrics.MetricCpuLimit.Name] = intValue(val.MilliValue())
	} else {
		metricSet.Values[metrics.MetricCpuLimit.Name] = intValue(0)
	}
	if val, found := limits[kube_api.ResourceMemory]; found {
		metricSet.Values[metrics.MetricMemoryLimit.Name] = intValue(val.Value())
	} else {
		metricSet.Values[metrics.MetricMemoryLimit.Name] = intValue(0)
	}
	if val, found := limits[kube_api.ResourceEphemeralStorage]; found {
		metricSet.Values[metrics.MetricEphemeralStorageLimit.Name] = intValue(val.Value())
	} else {
		metricSet.Values[metrics.MetricEphemeralStorageLimit.Name] = intValue(0)
	}
}

func (pbe *PodBasedEnricher) addContainerStatus(collectionTime time.Time, containerMs *metrics.Set, metric *metrics.Metric, status kube_api.ContainerStatus) {
	labels := make(map[string]string, 2)

	stateInt, state, reason, exitCode := pbe.findContainerState(collectionTime, status)

	if stateInt > 0 {
		labels["status"] = state
		if reason != "" {
			labels["reason"] = reason
			labels["exit_code"] = fmt.Sprint(exitCode)
		}
	}
	addLabeledIntMetric(containerMs, metric, labels, stateInt)
}

func (pbe *PodBasedEnricher) findContainerState(collectionTime time.Time, status kube_api.ContainerStatus) (int64, string, string, int32) {
	if status.LastTerminationState.Terminated == nil {
		return convertContainerState(status.State)
	}

	lastTerminationTime := status.LastTerminationState.Terminated.FinishedAt.Time
	lastCollectionTime := collectionTime.Add(-1 * pbe.collectionInterval)
	if lastCollectionTime.After(lastTerminationTime) {
		return convertContainerState(status.State)
	}

	return convertContainerState(status.LastTerminationState)
}

func (pbe *PodBasedEnricher) updateContainerStatus(metricSets map[metrics.ResourceKey]*metrics.Set, pod *kube_api.Pod, statuses []kube_api.ContainerStatus, collectionTime time.Time) {
	if len(statuses) == 0 {
		return
	}
	for _, status := range statuses {
		containerKey := metrics.PodContainerKey(pod.Namespace, pod.Name, status.Name)
		containerMs, found := metricSets[containerKey]
		if !found {
			log.Debugf("Container key %s not found", containerKey)
			continue
		}
		pbe.addContainerStatus(collectionTime, containerMs, &metrics.MetricContainerStatus, status)
	}
}

func convertPhase(phase kube_api.PodPhase) int64 {
	switch phase {
	case kube_api.PodPending:
		return 1
	case kube_api.PodRunning:
		return 2
	case kube_api.PodSucceeded:
		return 3
	case kube_api.PodFailed:
		return 4
	case kube_api.PodUnknown:
		return 5
	default:
		return 5
	}
}

func convertContainerState(state kube_api.ContainerState) (int64, string, string, int32) {
	if state.Running != nil {
		return 1, "running", "", 0
	}
	if state.Waiting != nil {
		return 2, "waiting", state.Waiting.Reason, 0
	}
	if state.Terminated != nil {
		return 3, "terminated", state.Terminated.Reason, state.Terminated.ExitCode
	}
	return 0, "", "", 0
}

// addLabeledIntMetric is a convenience method for adding the labeled metric and value to the metric set.
func addLabeledIntMetric(ms *metrics.Set, metric *metrics.Metric, labels map[string]string, value int64) {
	val := metrics.LabeledValue{
		Name:   metric.Name,
		Labels: labels,
		Value: metrics.Value{
			ValueType: metrics.ValueInt64,
			IntValue:  value,
		},
	}
	ms.LabeledValues = append(ms.LabeledValues, val)
}

func intValue(value int64) metrics.Value {
	return metrics.Value{
		IntValue:  value,
		ValueType: metrics.ValueInt64,
	}
}

func NewPodBasedEnricher(podLister v1listers.PodLister, labelCopier *util.LabelCopier, collectionInterval time.Duration) *PodBasedEnricher {
	return &PodBasedEnricher{
		podLister:          podLister,
		labelCopier:        labelCopier,
		collectionInterval: collectionInterval,
	}
}
