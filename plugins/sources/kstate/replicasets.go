// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"reflect"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	appsv1 "k8s.io/api/apps/v1"
)

func pointsForReplicaSet(item interface{}, transforms configuration.Transforms) []*wf.Point {
	rs, ok := item.(*appsv1.ReplicaSet)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	tags := buildTags("replicaset", rs.Name, rs.Namespace, transforms.Tags)
	now := time.Now().Unix()
	desired := floatVal(rs.Spec.Replicas, 1.0)
	available := float64(rs.Status.AvailableReplicas)
	ready := float64(rs.Status.ReadyReplicas)

	return []*wf.Point{
		metricPoint(transforms.Prefix, "replicaset.desired_replicas", desired, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "replicaset.available_replicas", available, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "replicaset.ready_replicas", ready, now, transforms.Source, tags),
	}
}
