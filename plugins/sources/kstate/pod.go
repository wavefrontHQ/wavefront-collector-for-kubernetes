package kstate

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"k8s.io/api/core/v1"
)

func buildPodPhase(pod *v1.Pod, prefix, source string, tags map[string]string, ts int64) *metrics.MetricPoint {
	return &metrics.MetricPoint{
		Metric:    prefix + "pod.status.phase",
		Value:     float64(convertPhase(pod.Status.Phase)),
		Timestamp: ts,
		Source:    source,
		Tags:      tags,
	}
}

func buildPodCondition(pod *v1.Pod) *metrics.MetricPoint {
	//kube_pod_status_ready	Gauge	pod=<pod-name>
	//	namespace=<pod-namespace>
	//	condition=<true|false|unknown>	STABLE
	//kube_pod_status_scheduled	Gauge	pod=<pod-name>
	//	namespace=<pod-namespace>
	//	condition=<true|false|unknown>	STABLE
	return nil
}

func buildContainerStatuses() {
	//TODO: implement
	//kube_pod_container_info	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>
	//	image=<image-name>
	//	image_id=<image-id>
	//	container_id=<containerid>	STABLE
	//kube_pod_container_status_waiting	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>	STABLE
	//kube_pod_container_status_waiting_reason	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>
	//	reason=<ContainerCreating|CrashLoopBackOff|ErrImagePull|ImagePullBackOff|CreateContainerConfigError|InvalidImageName|CreateContainerError>	STABLE
	//kube_pod_container_status_running	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>	STABLE
	//kube_pod_container_status_terminated	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>	STABLE
	//kube_pod_container_status_terminated_reason	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>
	//	reason=<OOMKilled|Error|Completed|ContainerCannotRun|DeadlineExceeded>	STABLE
	//kube_pod_container_status_last_terminated_reason	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>
	//	reason=<OOMKilled|Error|Completed|ContainerCannotRun|DeadlineExceeded>	STABLE
	//kube_pod_container_status_ready	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>	STABLE
	//kube_pod_container_status_restarts_total	Counter	container=<container-name>
	//	namespace=<pod-namespace>
	//	pod=<pod-name>

	//kube_pod_init_container_status_waiting	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>	STABLE
	//kube_pod_init_container_status_waiting_reason	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>
	//	reason=<ContainerCreating|CrashLoopBackOff|ErrImagePull|ImagePullBackOff|CreateContainerConfigError>	STABLE
	//kube_pod_init_container_status_running	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>	STABLE
	//kube_pod_init_container_status_terminated	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>	STABLE
	//kube_pod_init_container_status_terminated_reason	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>
	//	reason=<OOMKilled|Error|Completed|ContainerCannotRun|DeadlineExceeded>	STABLE
	//kube_pod_init_container_status_last_terminated_reason	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>
	//	reason=<OOMKilled|Error|Completed|ContainerCannotRun|DeadlineExceeded>	STABLE
	//kube_pod_init_container_status_ready	Gauge	container=<container-name>
	//	pod=<pod-name>
	//	namespace=<pod-namespace>	STABLE
	//kube_pod_init_container_status_restarts_total	Counter	container=<container-name>
	//	namespace=<pod-namespace>
	//	pod=<pod-name>

	//kube_pod_status_scheduled_time	Gauge	pod=<pod-name>
	//	namespace=<pod-namespace>	STABLE
	//kube_pod_status_unschedulable	Gauge	pod=<pod-name>
	//	namespace=<pod-namespace>
}

func convertPhase(phase v1.PodPhase) int64 {
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
