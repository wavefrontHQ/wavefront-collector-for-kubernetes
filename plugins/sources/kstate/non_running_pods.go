// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"reflect"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
	v1 "k8s.io/api/core/v1"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
)

func pointsForNonRunningPods(item interface{}, transforms configuration.Transforms) []*wf.Point {
	pod, ok := item.(*v1.Pod)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	sharedTags := make(map[string]string, len(pod.GetLabels())+1)
	copyLabels(pod.GetLabels(), sharedTags)
	now := time.Now().Unix()

	points := buildPodPhaseMetrics(pod, transforms, sharedTags, now)

	points = append(points, buildContainerStatusMetrics(pod, sharedTags, transforms, now)...)
	return points
}

func truncateMessage(message string) string {
	maxPointTagLength := 255 - len("=") - len("message")
	if len(message) >= maxPointTagLength {
		return message[0:maxPointTagLength]
	}
	return message
}

func buildPodPhaseMetrics(pod *v1.Pod, transforms configuration.Transforms, sharedTags map[string]string, now int64) []*wf.Point {
	tags := buildTags("pod_name", pod.Name, pod.Namespace, transforms.Tags)
	tags[metrics.LabelMetricSetType.Key] = metrics.MetricSetTypePod
	tags[metrics.LabelPodId.Key] = string(pod.UID)
	tags["phase"] = string(pod.Status.Phase)

	phaseValue := util.ConvertPodPhase(pod.Status.Phase)
	if phaseValue == util.POD_PHASE_PENDING {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodScheduled && condition.Status == "False" {
				tags[metrics.LabelNodename.Key] = "none"
				tags["reason"] = condition.Reason
				tags["message"] = truncateMessage(condition.Message)
			} else if condition.Type == v1.ContainersReady && condition.Status == "False" {
				tags["reason"] = condition.Reason
				tags["message"] = truncateMessage(condition.Message)
			}
		}
	}

	if phaseValue == util.POD_PHASE_FAILED {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady {
				tags["reason"] = condition.Reason
				tags["message"] = truncateMessage(condition.Message)
			}
		}
	}

	nodeName := pod.Spec.NodeName
	if len(nodeName) > 0 {
		sharedTags[metrics.LabelNodename.Key] = nodeName
	}
	copyTags(sharedTags, tags)
	points := []*wf.Point{
		metricPoint(transforms.Prefix, "pod.status.phase", float64(phaseValue), now, transforms.Source, tags),
	}
	return points
}

func buildContainerStatusMetrics(pod *v1.Pod, sharedTags map[string]string, transforms configuration.Transforms, now int64) []*wf.Point {
	statuses := pod.Status.ContainerStatuses
	if len(statuses) == 0 {
		return []*wf.Point{}
	}

	points := make([]*wf.Point, len(statuses))
	for i, status := range statuses {
		containerStateInfo := util.NewContainerStateInfo(status.State)
		tags := buildTags("pod_name", pod.Name, pod.Namespace, transforms.Tags)
		tags[metrics.LabelMetricSetType.Key] = metrics.MetricSetTypePodContainer
		tags[metrics.LabelContainerName.Key] = status.Name
		tags[metrics.LabelContainerBaseImage.Key] = status.Image

		copyTags(sharedTags, tags)
		containerStateInfo.AddMetricTags(tags)

		points[i] = metricPoint(transforms.Prefix, "pod_container.status", float64(containerStateInfo.Value), now, transforms.Source, tags)
	}
	return points
}
