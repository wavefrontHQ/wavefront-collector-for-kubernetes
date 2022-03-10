package kstate

import (
	"testing"

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
		assert.Equal(t, float64(1), actualWFPoints[0].Value)
		assert.Equal(t, "pod1", actualWFPoints[0].Tags()["pod_name"])
		assert.Equal(t, string(v1.PodPending), actualWFPoints[0].Tags()["phase"])
		assert.Equal(t, "testLabelName", actualWFPoints[0].Tags()["label.name"])
		assert.Equal(t, "Unschedulable", actualWFPoints[0].Tags()["reason"])
		assert.Equal(t, "0/1 nodes are available: 1 Insufficient memory.", actualWFPoints[0].Tags()["message"])
	})

	t.Run("test for completed pod", func(t *testing.T) {
		testPod := setupCompletedPod()
		actualWFPoints := pointsForNonRunningPods(testPod, testTransform)
		assert.Equal(t, 2, len(actualWFPoints))

		// check for pod metrics
		assert.Equal(t, float64(3), actualWFPoints[0].Value)
		assert.Equal(t, string(v1.PodSucceeded), actualWFPoints[0].Tags()["phase"])
		assert.Equal(t, "", actualWFPoints[0].Tags()["reason"])

		// check for container metrics
		assert.Equal(t, float64(3), actualWFPoints[1].Value)
		assert.Equal(t, "0", actualWFPoints[1].Tags()["exit_code"])
		assert.Equal(t, "Completed", actualWFPoints[1].Tags()["reason"])
		assert.Equal(t, "terminated", actualWFPoints[1].Tags()["status"])
	})

	t.Run("test for failed pod", func(t *testing.T) {
		testPod := setupFailedPod()
		actualWFPoints := pointsForNonRunningPods(testPod, testTransform)
		assert.Equal(t, 2, len(actualWFPoints))

		// check for pod metrics
		assert.Equal(t, float64(4), actualWFPoints[0].Value)
		assert.Equal(t, string(v1.PodFailed), actualWFPoints[0].Tags()["phase"])
		assert.Equal(t, "ContainersNotReady", actualWFPoints[0].Tags()["reason"])
		assert.Equal(t, 255, len(actualWFPoints[0].Tags()["message"])+len("message")+len("="))
		assert.Contains(t, actualWFPoints[0].Tags()["message"], "containers with unready status: [hello]")

		// check for container metrics
		assert.Equal(t, float64(3), actualWFPoints[1].Value)
		assert.Equal(t, "1", actualWFPoints[1].Tags()["exit_code"])
		assert.Equal(t, "Error", actualWFPoints[1].Tags()["reason"])
		assert.Equal(t, "terminated", actualWFPoints[1].Tags()["status"])
	})

	t.Run("test for container creating pod", func(t *testing.T) {
		testPod := setupContainerCreatingPod()
		actualWFPoints := pointsForNonRunningPods(testPod, testTransform)
		assert.Equal(t, 2, len(actualWFPoints))

		// check for pod metrics
		assert.Equal(t, float64(1), actualWFPoints[0].Value)
		assert.Equal(t, string(v1.PodPending), actualWFPoints[0].Tags()["phase"])
		assert.Equal(t, "ContainersNotReady", actualWFPoints[0].Tags()["reason"])
		assert.Equal(t, "containers with unready status: [wavefront-proxy]", actualWFPoints[0].Tags()["message"])

		// check for container metrics
		assert.Equal(t, float64(2), actualWFPoints[1].Value)
		assert.Equal(t, "ContainerCreating", actualWFPoints[1].Tags()["reason"])
		assert.Equal(t, "waiting", actualWFPoints[1].Tags()["status"])
	})
}
