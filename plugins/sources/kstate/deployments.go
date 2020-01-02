// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	log "github.com/sirupsen/logrus"
	"reflect"
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	appsv1 "k8s.io/api/apps/v1"
)

func pointsForDeployment(item interface{}, transforms configuration.Transforms) []*metrics.MetricPoint {
	deployment, ok := item.(*appsv1.Deployment)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	tags := buildTags("deployment", deployment.Name, deployment.Namespace, transforms.Tags)
	now := time.Now().Unix()
	available := float64(deployment.Status.AvailableReplicas)
	desired := floatVal(deployment.Spec.Replicas, 0.0)

	return []*metrics.MetricPoint{
		metricPoint(transforms.Prefix, "deployment.available_replicas", available, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "deployment.desired_replicas", desired, now, transforms.Source, tags),
	}
}
