// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import (
	"reflect"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
)

func pointsForCronJob(item interface{}, transforms configuration.Transforms) []wf.Metric {
	job, ok := item.(*batchv1beta1.CronJob)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	tags := buildTags("cronjob", job.Name, job.Namespace, transforms.Tags)
	now := time.Now().Unix()
	active := float64(len(job.Status.Active))

	return []wf.Metric{
		metricPoint(transforms.Prefix, "cronjob.active", active, now, transforms.Source, tags),
	}
}
