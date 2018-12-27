package discovery

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FakeManager struct {
	registeredPods map[string]string
}

func NewFakeManager() *FakeManager {
	return &FakeManager{
		registeredPods: make(map[string]string),
	}
}

func (f *FakeManager) ListPods(ns string, labels map[string]string) ([]*v1.Pod, error) {
	pods := make([]*v1.Pod, 2)
	pods[0] = FakePod("pod1", "ns", "123")
	pods[1] = FakePod("pod2", "ns", "124")
	return pods, nil
}

func (f *FakeManager) Registered(name string) string {
	return f.registeredPods[name]
}

func (f *FakeManager) RegisterProvider(podName string, provider metrics.MetricsSourceProvider, obj string) {
	f.registeredPods[podName] = obj
}

func (f *FakeManager) UnregisterProvider(podName, providerName string) {
	delete(f.registeredPods, podName)
}

func FakePod(name, namespace, ip string) *v1.Pod {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: v1.PodStatus{
			PodIP: ip,
		},
	}
	return &pod
}
