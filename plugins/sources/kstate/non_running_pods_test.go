package kstate

import (
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setupBasicPod() *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "ns1",
			Labels:    map[string]string{"name": "testLabelName"},
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
		},
	}
	return pod
}

func setupPendingPod() *v1.Pod {
	pendingPod := setupBasicPod()
	pendingPod.Spec.NodeName = ""
	pendingPod.Status = v1.PodStatus{
		Phase: v1.PodPending,
		Conditions: []v1.PodCondition{
			{
				Type:    "PodScheduled",
				Status:  "False",
				Reason:  "Unschedulable",
				Message: "0/1 nodes are available: 1 Insufficient memory.",
			},
		},
	}
	return pendingPod
}

func setupContainerCreatingPod() *v1.Pod {
	containerCreatingPod := setupBasicPod()
	containerCreatingPod.Status = v1.PodStatus{
		Phase: v1.PodPending,
		Conditions: []v1.PodCondition{
			{
				Type:   "Initialized",
				Status: "True",
			},
			{
				Type:    "Ready",
				Status:  "False",
				Reason:  "ContainersNotReady",
				Message: "containers with unready status: [wavefront-proxy]",
			},
			{
				Type:    "ContainersReady",
				Status:  "False",
				Reason:  "ContainersNotReady",
				Message: "containers with unready status: [wavefront-proxy]",
			},
			{
				Type:   "PodScheduled",
				Status: "True",
			},
		},
		ContainerStatuses: []v1.ContainerStatus{
			{
				Name: "testContainerName",
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason: "ContainerCreating",
					},
				},
				Ready:   false,
				Image:   "testImage",
				ImageID: "",
			},
		},
	}
	return containerCreatingPod
}

func setupCompletedPod() *v1.Pod {
	completedPod := setupBasicPod()
	completedPod.Status = v1.PodStatus{
		Phase: v1.PodSucceeded,
		Conditions: []v1.PodCondition{
			{
				Type:   "Initialized",
				Status: "True",
				Reason: "PodCompleted",
			},
			{
				Type:   "Ready",
				Status: "False",
				Reason: "PodCompleted",
			},
			{
				Type:   "ContainersReady",
				Status: "False",
				Reason: "PodCompleted",
			},
			{
				Type:   "PodScheduled",
				Status: "True",
			},
		},
		ContainerStatuses: []v1.ContainerStatus{
			{
				Name: "testContainerName",
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode:    0,
						Reason:      "Completed",
						ContainerID: "testContainerID",
					},
				},
				Ready:   false,
				Image:   "testImage",
				ImageID: "testImageID",
			},
		},
	}
	return completedPod
}

func setupFailedPod() *v1.Pod {
	failedPod := setupBasicPod()
	failedPod.Status = v1.PodStatus{
		Phase: v1.PodFailed,
		Conditions: []v1.PodCondition{
			{
				Type:   "Initialized",
				Status: "True",
			},
			{
				Type:    "Ready",
				Status:  "False",
				Reason:  "ContainersNotReady",
				Message: "containers with unready status: [hello], and this message exceeds 255 characters point tag. Maximum allowed length for a combination of a point tag key and value is 254 characters (255 including the = separating key and value). If the value is longer, the point is rejected and logged. Keep the number of distinct time series per metric and host to under 1000.",
			},
			{
				Type:    "ContainersReady",
				Status:  "False",
				Reason:  "ContainersNotReady",
				Message: "containers with unready status: [hello], and this message exceeds 255 characters point tag. Maximum allowed length for a combination of a point tag key and value is 254 characters (255 including the = separating key and value). If the value is longer, the point is rejected and logged. Keep the number of distinct time series per metric and host to under 1000.",
			},
			{
				Type:   "PodScheduled",
				Status: "True",
			},
		},
		ContainerStatuses: []v1.ContainerStatus{
			{
				Name: "testContainerName",
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode:    1,
						Reason:      "Error",
						ContainerID: "testContainerID",
					},
				},
				Ready:   false,
				Image:   "testImage",
				ImageID: "testImageID",
			},
		},
	}
	return failedPod
}

func setupTestTransform() configuration.Transforms {
	return configuration.Transforms{
		Source:  "testSource",
		Prefix:  "testPrefix",
		Tags:    nil,
		Filters: filter.Config{},
	}
}

func TestPointsForNonRunningPods(t *testing.T) {
	testTransform := setupTestTransform()

	t.Run("test for pending pod", func(t *testing.T) {
		testPod := setupPendingPod()
		actualWFPoints := pointsForNonRunningPods(testPod, testTransform)
		assert.Equal(t, 1, len(actualWFPoints))
		point := actualWFPoints[0].(*wf.Point)
		assert.Equal(t, float64(util.POD_PHASE_PENDING), point.Value)
		assert.Equal(t, "pod1", point.Tags()["pod_name"])
		assert.Equal(t, string(v1.PodPending), point.Tags()["phase"])
		assert.Equal(t, "testLabelName", point.Tags()["label.name"])
		assert.Equal(t, "Unschedulable", point.Tags()["reason"])
		assert.Equal(t, "none", point.Tags()["nodename"])
		assert.Equal(t, "0/1 nodes are available: 1 Insufficient memory.", point.Tags()["message"])
	})

	t.Run("test for completed pod", func(t *testing.T) {
		testPod := setupCompletedPod()
		actualWFPoints := pointsForNonRunningPods(testPod, testTransform)
		assert.Equal(t, 2, len(actualWFPoints))

		// check for pod metrics
		podPoint := actualWFPoints[0].(*wf.Point)
		assert.Equal(t, float64(util.POD_PHASE_SUCCEEDED), podPoint.Value)
		assert.Equal(t, string(v1.PodSucceeded), podPoint.Tags()["phase"])
		assert.Equal(t, "", podPoint.Tags()["reason"])
		assert.Equal(t, "node1", podPoint.Tags()["nodename"])

		// check for container metrics
		containerPoint := actualWFPoints[1].(*wf.Point)
		assert.Equal(t, float64(util.CONTAINER_STATE_TERMINATED), containerPoint.Value)
		assert.Equal(t, "0", containerPoint.Tags()["exit_code"])
		assert.Equal(t, "Completed", containerPoint.Tags()["reason"])
		assert.Equal(t, "terminated", containerPoint.Tags()["status"])
	})

	t.Run("test for failed pod", func(t *testing.T) {
		testPod := setupFailedPod()
		actualWFPoints := pointsForNonRunningPods(testPod, testTransform)
		assert.Equal(t, 2, len(actualWFPoints))

		// check for pod metrics
		podPoint := actualWFPoints[0].(*wf.Point)
		assert.Equal(t, float64(util.POD_PHASE_FAILED), podPoint.Value)
		assert.Equal(t, string(v1.PodFailed), podPoint.Tags()["phase"])
		assert.Equal(t, "ContainersNotReady", podPoint.Tags()["reason"])
		assert.Equal(t, 255, len(podPoint.Tags()["message"])+len("message")+len("="))
		assert.Contains(t, podPoint.Tags()["message"], "containers with unready status: [hello]")
		assert.Equal(t, "node1", podPoint.Tags()["nodename"])

		// check for container metrics
		containerMetric := actualWFPoints[1].(*wf.Point)
		assert.Equal(t, float64(util.CONTAINER_STATE_TERMINATED), containerMetric.Value)
		assert.Equal(t, "1", containerMetric.Tags()["exit_code"])
		assert.Equal(t, "Error", containerMetric.Tags()["reason"])
		assert.Equal(t, "terminated", containerMetric.Tags()["status"])
	})

	t.Run("test for container creating pod", func(t *testing.T) {
		testPod := setupContainerCreatingPod()
		actualWFPoints := pointsForNonRunningPods(testPod, testTransform)
		assert.Equal(t, 2, len(actualWFPoints))

		// check for pod metrics
		podMetric := actualWFPoints[0].(*wf.Point)
		assert.Equal(t, float64(util.POD_PHASE_PENDING), podMetric.Value)
		assert.Equal(t, string(v1.PodPending), podMetric.Tags()["phase"])
		assert.Equal(t, "ContainersNotReady", podMetric.Tags()["reason"])
		assert.Equal(t, "containers with unready status: [wavefront-proxy]", podMetric.Tags()["message"])
		assert.Equal(t, "node1", podMetric.Tags()["nodename"])

		// check for container metrics
		containerMetric := actualWFPoints[1].(*wf.Point)
		assert.Equal(t, float64(util.CONTAINER_STATE_WAITING), containerMetric.Value)
		assert.Equal(t, "ContainerCreating", containerMetric.Tags()["reason"])
		assert.Equal(t, "waiting", containerMetric.Tags()["status"])
	})
}
