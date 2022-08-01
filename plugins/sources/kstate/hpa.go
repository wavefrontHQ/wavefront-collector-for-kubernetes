// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"reflect"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"
	"k8s.io/api/autoscaling/v2beta2"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
)

func pointsForHPA(item interface{}, transforms configuration.Transforms) []wf.Metric {
	hpa, ok := item.(*v2beta2.HorizontalPodAutoscaler)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	tags := buildTags("hpa", hpa.Name, hpa.Namespace, transforms.Tags)
	now := time.Now().Unix()
	maxReplicas := float64(hpa.Spec.MaxReplicas)
	minReplicas := floatVal(hpa.Spec.MinReplicas, 1.0)
	currReplicas := float64(hpa.Status.CurrentReplicas)
	desiredReplicas := float64(hpa.Status.DesiredReplicas)

	return []wf.Metric{
		metricPoint(transforms.Prefix, "hpa.max_replicas", maxReplicas, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "hpa.min_replicas", minReplicas, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "hpa.current_replicas", currReplicas, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "hpa.desired_replicas", desiredReplicas, now, transforms.Source, tags),
	}
}
