package discovery

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FakeResourceLister struct {
	count          int
	registeredPods map[string]string
}

func NewFakeResourceLister(count int) *FakeResourceLister {
	return &FakeResourceLister{
		count:          count,
		registeredPods: make(map[string]string),
	}
}

func (f *FakeResourceLister) ListPods(ns string, labels map[string]string) ([]*v1.Pod, error) {
	pods := make([]*v1.Pod, f.count)
	for i := 0; i < f.count; i++ {
		pods[i] = FakePod("pod"+string(i), "ns", "192.168.0.123")
	}
	return pods, nil
}

func (f *FakeResourceLister) ListServices(ns string, labels map[string]string) ([]*v1.Service, error) {
	services := make([]*v1.Service, f.count)
	for i := 0; i < f.count; i++ {
		services[i] = FakeService("svc"+string(i), "ns", "192.168.0.123")
	}
	return services, nil
}

func FakeService(name, namespace, ip string) *v1.Service {
	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: ip,
		},
	}
	return &service
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
