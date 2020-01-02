package kstate

import (
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"k8s.io/api/autoscaling/v2beta1"
)

func pointsForHPA(item interface{}, transforms configuration.Transforms) []*metrics.MetricPoint {
	hpa, ok := item.(*v2beta1.HorizontalPodAutoscaler)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}

	tags := buildTags("hpa", hpa.Name, hpa.Namespace, transforms.Tags)
	now := time.Now().Unix()
	maxReplicas := float64(hpa.Spec.MaxReplicas)
	minReplicas := floatVal(hpa.Spec.MinReplicas, 0.0)
	currReplicas := float64(hpa.Status.CurrentReplicas)
	desiredReplicas := float64(hpa.Status.DesiredReplicas)

	return []*metrics.MetricPoint{
		metricPoint(transforms.Prefix, "hpa.max_replicas", maxReplicas, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "hpa.min_replicas", minReplicas, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "hpa.current_replicas", currReplicas, now, transforms.Source, tags),
		metricPoint(transforms.Prefix, "hpa.desired_replicas", desiredReplicas, now, transforms.Source, tags),
	}
}
