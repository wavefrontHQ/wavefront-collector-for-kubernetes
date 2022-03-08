// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
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


	tags := buildTags("pod", pod.Name, pod.Namespace, transforms.Tags)
    tags[metrics.LabelMetricSetType.Key] = metrics.MetricSetTypePod
    tags[metrics.LabelPodId.Key] = string(pod.UID)
    tags[metrics.LabelPodName.Key] = pod.Name
    tags["phase"] = string(pod.Status.Phase)
    copyLabels(pod.GetLabels(), tags)

    now := time.Now().Unix()

	phaseValue :=convertPhase(pod.Status.Phase)
	return []*wf.Point{
		metricPoint(transforms.Prefix, "pod.status.phase", float64(phaseValue), now, transforms.Source, tags),
	}
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
