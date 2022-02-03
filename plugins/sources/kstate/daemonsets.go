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

func pointsForDaemonSet(item interface{}, transforms configuration.Transforms) []*wf.Point {
	ds, ok := item.(*appsv1.DaemonSet)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	tags := buildTags("daemonset", ds.Name, ds.Namespace, transforms.Tags)
	now := time.Now().Unix()
	currentScheduled := float64(ds.Status.CurrentNumberScheduled)
	desiredScheduled := float64(ds.Status.DesiredNumberScheduled)
	misScheduled := float64(ds.Status.NumberMisscheduled)
	ready := float64(ds.Status.NumberReady)

	return []*wf.Point{
		metricPoint(transforms.Prefix, "daemonset.current_scheduled", currentScheduled, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "daemonset.desired_scheduled", desiredScheduled, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "daemonset.misscheduled", misScheduled, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "daemonset.ready", ready, now, transforms.Source, tags),
	}
}
