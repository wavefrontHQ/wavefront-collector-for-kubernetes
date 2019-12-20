package kstate

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"strconv"

	"k8s.io/api/core/v1"
)

func buildPodPhase(pod *v1.Pod, prefix, source string, tags map[string]string, ts int64) *metrics.MetricPoint {
	pt := make(map[string]string, len(tags)+1)
	for k, v := range tags {
		pt[k] = v
	}
	pt["phase"] = string(pod.Status.Phase)
	return metricPoint(prefix+"pod.status.phase", convertPhase(pod.Status.Phase), ts, source, pt)
}

func buildContainerStatuses(statuses []v1.ContainerStatus, prefix, source string, tags map[string]string, ts int64) []*metrics.MetricPoint {
	if len(statuses) == 0 {
		return nil
	}

	points := make([]*metrics.MetricPoint, 2*len(statuses))
	for i, container := range statuses {
		pt := make(map[string]string, len(tags)+4)
		for k, v := range tags {
			pt[k] = v
		}

		pt["container_name"] = container.Name
		pt["container_image_name"] = container.Image
		pt["ready"] = strconv.FormatBool(container.Ready)

		stateFloat, state, reason := convertContainerStatus(container.State)
		if stateFloat > 0 {
			pt["status"] = state
			if reason != "" {
				pt["reason"] = reason
			}
		}

		idx := i * 2

		// status
		points[idx] = metricPoint(prefix+"status", stateFloat, ts, source, pt)

		// restart.count
		count := float64(container.RestartCount)
		points[idx+1] = metricPoint(prefix+"restart.count", count, ts, source, pt)
	}
	return points
}

func convertPhase(phase v1.PodPhase) float64 {
	switch phase {
	case v1.PodPending:
		return 1
	case v1.PodRunning:
		return 2
	case v1.PodSucceeded:
		return 3
	case v1.PodFailed:
		return 4
	case v1.PodUnknown:
		return 5
	default:
		return 5
	}
}

func convertContainerStatus(state v1.ContainerState) (float64, string, string) {
	if state.Running != nil {
		return 1, "running", ""
	}
	if state.Waiting != nil {
		return 2, "waiting", state.Waiting.Reason
	}
	if state.Terminated != nil {
		return 3, "terminated", state.Terminated.Reason
	}
	return 0, "", ""
}

func metricPoint(name string, value float64, ts int64, source string, tags map[string]string) *metrics.MetricPoint {
	return &metrics.MetricPoint{
		Metric:    name,
		Value:     value,
		Timestamp: ts,
		Source:    source,
		Tags:      tags,
	}
}
