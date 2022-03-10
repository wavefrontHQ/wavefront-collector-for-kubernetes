// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"fmt"
	"reflect"
	"time"

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

func buildPodPhaseMetrics(pod *v1.Pod, transforms configuration.Transforms, sharedTags map[string]string, now int64) []*wf.Point {
	tags := buildTags("pod_name", pod.Name, pod.Namespace, transforms.Tags)
	tags[metrics.LabelMetricSetType.Key] = metrics.MetricSetTypePod
	tags[metrics.LabelPodId.Key] = string(pod.UID)
	tags["phase"] = string(pod.Status.Phase)

	phaseValue := convertPhase(pod.Status.Phase)
	if phaseValue == 1 {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodScheduled && condition.Status == "False" {
				tags["reason"] = condition.Reason
			} else if condition.Type == v1.ContainersReady && condition.Status == "False" {
                tags["reason"] = condition.Reason
            }
		}
	}

	if phaseValue == 4 {
        log.Infof("failed phase")
		for _, condition := range pod.Status.Conditions {
            if condition.Type == v1.PodReady {
				tags["reason"] = condition.Reason
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

func convertPhase(phase v1.PodPhase) int64 {
	switch phase {
	case v1.PodPending:
		return 1
	case v1.PodRunning:
		return 2
	case v1.PodSucceeded:
		return 3
	case v1.PodFailed:
		return 4
	case v1.PodUnknown:
		return 5
	default:
		return 5
	}
}

func buildContainerStatusMetrics(pod *v1.Pod, sharedTags map[string]string, transforms configuration.Transforms, now int64) []*wf.Point {
	statuses := pod.Status.ContainerStatuses
	if len(statuses) == 0 {
		return []*wf.Point{}
	}

	//container_base_image
	points := make([]*wf.Point, len(statuses))
	for i, status := range statuses {
		stateInt, state, reason, exitCode := convertContainerState(status.State)
		tags := buildTags("pod_name", pod.Name, pod.Namespace, transforms.Tags)
		tags[metrics.LabelMetricSetType.Key] = metrics.MetricSetTypePodContainer
		tags[metrics.LabelContainerName.Key] = status.Name
		tags[metrics.LabelContainerBaseImage.Key] = status.Image

		copyTags(sharedTags, tags)
		if stateInt > 0 {
			tags["status"] = state
			if reason != "" {
				tags["reason"] = reason
				tags["exit_code"] = fmt.Sprint(exitCode)
			}
		}
		points[i] = metricPoint(transforms.Prefix, "pod_container.status", float64(stateInt), now, transforms.Source, tags)
	}
	return points
}

func convertContainerState(state v1.ContainerState) (int64, string, string, int32) {
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
