// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	batchv1 "k8s.io/api/batch/v1"
)

func pointsForJob(item interface{}, transforms configuration.Transforms) []*metrics.MetricPoint {
	job, ok := item.(*batchv1.Job)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	tags := buildTags("job", job.Name, job.Namespace, transforms.Tags)
	now := time.Now().Unix()
	active := float64(job.Status.Active)
	failed := float64(job.Status.Failed)
	succeeded := float64(job.Status.Succeeded)
	completions := floatVal(job.Spec.Completions, -1.0)
	parallelism := floatVal(job.Spec.Parallelism, -1.0)

	return []*metrics.MetricPoint{
		metricPoint(transforms.Prefix, "job.active", active, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "job.failed", failed, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "job.succeeded", succeeded, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "job.completions", completions, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "job.parallelism", parallelism, now, transforms.Source, tags),
	}
}
