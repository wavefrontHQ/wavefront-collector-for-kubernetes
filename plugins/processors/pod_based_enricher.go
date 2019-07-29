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

package processors

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/util"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	kube_api "k8s.io/api/core/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
)

type PodBasedEnricher struct {
	podLister   v1listers.PodLister
	labelCopier *util.LabelCopier
}

func (this *PodBasedEnricher) Name() string {
	return "pod_based_enricher"
}

func (this *PodBasedEnricher) Process(batch *metrics.DataBatch) (*metrics.DataBatch, error) {
	newMs := make(map[string]*metrics.MetricSet, len(batch.MetricSets))
	for k, v := range batch.MetricSets {
		switch v.Labels[metrics.LabelMetricSetType.Key] {
		case metrics.MetricSetTypePod:
			namespace := v.Labels[metrics.LabelNamespaceName.Key]
			podName := v.Labels[metrics.LabelPodName.Key]
			pod, err := this.getPod(namespace, podName)
			if err != nil {
				log.Debugf("Failed to get pod %s from cache: %v", metrics.PodKey(namespace, podName), err)
				continue
			}
			this.addPodInfo(k, v, pod, batch, newMs)
		case metrics.MetricSetTypePodContainer:
			namespace := v.Labels[metrics.LabelNamespaceName.Key]
			podName := v.Labels[metrics.LabelPodName.Key]
			pod, err := this.getPod(namespace, podName)
			if err != nil {
				log.Debugf("Failed to get pod %s from cache: %v", metrics.PodKey(namespace, podName), err)
				continue
			}
			this.addContainerInfo(k, v, pod, batch, newMs)
		}
	}
	for k, v := range newMs {
		batch.MetricSets[k] = v
	}
	return batch, nil
}

func (this *PodBasedEnricher) getPod(namespace, name string) (*kube_api.Pod, error) {
	pod, err := this.podLister.Pods(namespace).Get(name)
	if err != nil {
		return nil, err
	}
	if pod == nil {
		return nil, fmt.Errorf("cannot find pod definition")
	}
	return pod, nil
}

func (this *PodBasedEnricher) addContainerInfo(key string, containerMs *metrics.MetricSet, pod *kube_api.Pod, batch *metrics.DataBatch, newMs map[string]*metrics.MetricSet) {
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
			containerMs.MetricValues[metrics.MetricRestartCount.Name] = intValue(int64(containerStatus.RestartCount))
			if !pod.Status.StartTime.IsZero() {
				containerMs.EntityCreateTime = pod.Status.StartTime.Time
			}
			break
		}
	}

	containerMs.Labels[metrics.LabelPodId.Key] = string(pod.UID)
	this.labelCopier.Copy(pod.Labels, containerMs.Labels)

	namespace := containerMs.Labels[metrics.LabelNamespaceName.Key]
	podName := containerMs.Labels[metrics.LabelPodName.Key]

	podKey := metrics.PodKey(namespace, podName)
	_, oldfound := batch.MetricSets[podKey]
	if !oldfound {
		_, newfound := newMs[podKey]
		if !newfound {
			log.Debugf("Pod %s not found, creating a stub", podKey)
			podMs := &metrics.MetricSet{
				MetricValues: make(map[string]metrics.MetricValue),
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
			this.addPodInfo(podKey, podMs, pod, batch, newMs)
		}
	}
}

func (this *PodBasedEnricher) addPodInfo(key string, podMs *metrics.MetricSet, pod *kube_api.Pod, batch *metrics.DataBatch, newMs map[string]*metrics.MetricSet) {

	// Add UID and create time to pod
	podMs.Labels[metrics.LabelPodId.Key] = string(pod.UID)
	if !pod.Status.StartTime.IsZero() {
		podMs.EntityCreateTime = pod.Status.StartTime.Time
	}
	this.labelCopier.Copy(pod.Labels, podMs.Labels)

	// Add cpu/mem requests and limits to containers
	for _, container := range pod.Spec.Containers {
		containerKey := metrics.PodContainerKey(pod.Namespace, pod.Name, container.Name)
		if _, found := batch.MetricSets[containerKey]; found {
			continue
		}
		if _, found := newMs[containerKey]; found {
			continue
		}
		log.Debugf("Container %s not found, creating a stub", containerKey)
		containerMs := &metrics.MetricSet{
			MetricValues: make(map[string]metrics.MetricValue),
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
		this.labelCopier.Copy(pod.Labels, containerMs.Labels)
		updateContainerResourcesAndLimits(containerMs, container)
		newMs[containerKey] = containerMs
	}
}

func updateContainerResourcesAndLimits(metricSet *metrics.MetricSet, container kube_api.Container) {
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
					Type:        metrics.MetricGauge,
					ValueType:   metrics.ValueInt64,
					Units:       metrics.UnitsCount,
				},
			}
			metrics.ResourceRequestMetrics[key] = metric
		}
		if key == kube_api.ResourceCPU {
			metricSet.MetricValues[metric.Name] = intValue(val.MilliValue())
		} else {
			metricSet.MetricValues[metric.Name] = intValue(val.Value())
		}
	}

	// For primary resources like cpu and memory, explicitly sets their request resource
	// metric to zero if they are not requested.
	if _, found := requests[kube_api.ResourceCPU]; !found {
		metricSet.MetricValues[metrics.MetricCpuRequest.Name] = intValue(0)
	}
	if _, found := requests[kube_api.ResourceMemory]; !found {
		metricSet.MetricValues[metrics.MetricMemoryRequest.Name] = intValue(0)
	}
	if _, found := requests[kube_api.ResourceEphemeralStorage]; !found {
		metricSet.MetricValues[metrics.MetricEphemeralStorageRequest.Name] = intValue(0)
	}

	limits := container.Resources.Limits
	if val, found := limits[kube_api.ResourceCPU]; found {
		metricSet.MetricValues[metrics.MetricCpuLimit.Name] = intValue(val.MilliValue())
	} else {
		metricSet.MetricValues[metrics.MetricCpuLimit.Name] = intValue(0)
	}
	if val, found := limits[kube_api.ResourceMemory]; found {
		metricSet.MetricValues[metrics.MetricMemoryLimit.Name] = intValue(val.Value())
	} else {
		metricSet.MetricValues[metrics.MetricMemoryLimit.Name] = intValue(0)
	}
	if val, found := limits[kube_api.ResourceEphemeralStorage]; found {
		metricSet.MetricValues[metrics.MetricEphemeralStorageLimit.Name] = intValue(val.Value())
	} else {
		metricSet.MetricValues[metrics.MetricEphemeralStorageLimit.Name] = intValue(0)
	}
}

func intValue(value int64) metrics.MetricValue {
	return metrics.MetricValue{
		IntValue:   value,
		MetricType: metrics.MetricGauge,
		ValueType:  metrics.ValueInt64,
	}
}

func NewPodBasedEnricher(podLister v1listers.PodLister, labelCopier *util.LabelCopier) (*PodBasedEnricher, error) {
	return &PodBasedEnricher{
		podLister:   podLister,
		labelCopier: labelCopier,
	}, nil
}
