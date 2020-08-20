package kstate

import (
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

	"k8s.io/api/core/v1"
)

func pointsForNode(item interface{}, transforms configuration.Transforms) []*metrics.MetricPoint {
	node, ok := item.(*v1.Node)
	if !ok {
		log.Errorf("invalid type: %s", reflect.TypeOf(item).String())
		return nil
	}
	now := time.Now().Unix()
	points := buildNodeConditions(node, transforms, now)
	points = append(points, buildNodeTaints(node, transforms, now)...)
	return points
}

func buildNodeConditions(node *v1.Node, transforms configuration.Transforms, ts int64) []*metrics.MetricPoint {
	points := make([]*metrics.MetricPoint, len(node.Status.Conditions))
	for i, condition := range node.Status.Conditions {
		tags := buildTags("nodename", node.Name, "", transforms.Tags)
		copyLabels(node.GetLabels(), tags)
		tags["status"] = string(condition.Status)
		tags["condition"] = string(condition.Type)

		// add status and condition (condition.status and condition.type)
		points[i] = metricPoint(transforms.Prefix, "node.status.condition",
			nodeConditionFloat64(condition.Status), ts, transforms.Source, tags)
	}
	return points
}

func buildNodeTaints(node *v1.Node, transforms configuration.Transforms, ts int64) []*metrics.MetricPoint {
	points := make([]*metrics.MetricPoint, len(node.Spec.Taints))
	for i, taint := range node.Spec.Taints {
		tags := buildTags("nodename", node.Name, "", transforms.Tags)
		copyLabels(node.GetLabels(), tags)
		tags["key"] = taint.Key
		tags["value"] = taint.Value
		tags["effect"] = string(taint.Effect)
		points[i] = metricPoint(transforms.Prefix, "node.spec.taint", 1.0, ts, transforms.Source, tags)
	}
	return points
}

func nodeConditionFloat64(status v1.ConditionStatus) float64 {
	switch status {
	case v1.ConditionTrue:
		return 1.0
	case v1.ConditionFalse:
		return 0.0
	default:
		return -1.0
	}
}
