// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

	batchv1beta1 "k8s.io/api/batch/v1beta1"
)

func pointsForCronJob(item interface{}, transforms configuration.Transforms) []*metrics.MetricPointWithTags {
	job, ok := item.(*batchv1beta1.CronJob)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	tags := buildTags("cronjob", job.Name, job.Namespace, transforms.Tags)
	now := time.Now().Unix()
	active := float64(len(job.Status.Active))

	return []*metrics.MetricPointWithTags{
		metricPoint(transforms.Prefix, "cronjob.active", active, now, transforms.Source, tags),
	}
}
