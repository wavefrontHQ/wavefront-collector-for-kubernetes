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
	points := make([]*metrics.MetricPoint, 0)
	for _, condition := range node.Status.Conditions {
		tags := buildTags("nodename", node.Name, "", transforms.Tags)
		copyLabels(node.GetLabels(), tags)
		tags["status"] = string(condition.Status)
		tags["condition"] = string(condition.Type)

		// add status and condition (condition.status and condition.type)
		point := metricPoint(transforms.Prefix, "node.status.condition", nodeConditionFloat64(condition.Status), now, transforms.Source, tags)
		points = append(points, point)
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
