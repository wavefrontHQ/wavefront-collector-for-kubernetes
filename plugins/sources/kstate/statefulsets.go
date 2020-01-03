// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	appsv1 "k8s.io/api/apps/v1"
)

func pointsForStatefulSet(item interface{}, transforms configuration.Transforms) []*metrics.MetricPoint {
	ss, ok := item.(*appsv1.StatefulSet)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	tags := buildTags("statefulset", ss.Name, ss.Namespace, transforms.Tags)
	now := time.Now().Unix()

	desired := floatVal(ss.Spec.Replicas, 1.0)
	ready := float64(ss.Status.ReadyReplicas)
	current := float64(ss.Status.CurrentReplicas)
	updated := float64(ss.Status.UpdatedReplicas)

	return []*metrics.MetricPoint{
		metricPoint(transforms.Prefix, "statefulset.desired_replicas", desired, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "statefulset.current_replicas", current, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "statefulset.ready_replicas", ready, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "statefulset.updated_replicas", updated, now, transforms.Source, tags),
	}
}
