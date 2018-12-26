package discovery

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"k8s.io/api/core/v1"
)

type Manager interface {
	ListPods(ns string, labels map[string]string) ([]*v1.Pod, error)
	Registered(name string) string
	RegisterProvider(podName string, provider metrics.MetricsSourceProvider, obj string)
	UnregisterProvider(podName, providerName string)
}

type Discoverer interface {
	Discover(pod *v1.Pod) error
	Delete(pod *v1.Pod)
	Process(config Config) error
}
