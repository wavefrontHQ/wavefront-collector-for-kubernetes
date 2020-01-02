package kstate

import (
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	batchv1beta1 "k8s.io/api/batch/v1beta1"
)

func pointsForCronJob(item interface{}, transforms configuration.Transforms) []*metrics.MetricPoint {
	job, ok := item.(*batchv1beta1.CronJob)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	tags := buildTags("cronjob", job.Name, job.Namespace, transforms.Tags)
	now := time.Now().Unix()
	active := float64(len(job.Status.Active))

	return []*metrics.MetricPoint{
		metricPoint(transforms.Prefix, "cronjob.active", active, now, transforms.Source, tags),
	}
}
