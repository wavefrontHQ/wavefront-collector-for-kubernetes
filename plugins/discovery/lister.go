package discovery

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1listers "k8s.io/client-go/listers/core/v1"
)

type resourceLister struct {
	podLister     v1listers.PodLister
	serviceLister v1listers.ServiceLister
}

func newResourceLister(pl v1listers.PodLister, sl v1listers.ServiceLister) discovery.ResourceLister {
	return &resourceLister{
		podLister:     pl,
		serviceLister: sl,
	}
}

func (rl *resourceLister) ListPods(ns string, l map[string]string) ([]*apiv1.Pod, error) {
	if ns == "" {
		return rl.podLister.List(labels.SelectorFromSet(l))
	}
	nsLister := rl.podLister.Pods(ns)
	return nsLister.List(labels.SelectorFromSet(l))
}

func (rl *resourceLister) ListServices(ns string, l map[string]string) ([]*apiv1.Service, error) {
	if ns == "" {
		return rl.serviceLister.List(labels.SelectorFromSet(l))
	}
	nsLister := rl.serviceLister.Services(ns)
	return nsLister.List(labels.SelectorFromSet(l))
}
