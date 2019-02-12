package discovery

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceType int

const (
	PodType       ResourceType = 1
	ServiceType   ResourceType = 2
	IngressType   ResourceType = 3
	ApiServerType ResourceType = 4
)

func (resType ResourceType) String() string {
	switch resType {
	case PodType:
		return "pod"
	case ServiceType:
		return "service"
	case IngressType:
		return "ingress"
	case ApiServerType:
		return "apiserver"
	default:
		return fmt.Sprintf("%d", int(resType))
	}
}

type ResourceLister interface {
	ListPods(ns string, labels map[string]string) ([]*v1.Pod, error)
	ListServices(ns string, labels map[string]string) ([]*v1.Service, error)
}

type Discoverer interface {
	Discover(ip, kind string, meta metav1.ObjectMeta)
	Delete(kind string, meta metav1.ObjectMeta)
}

// Handles a single discovery rule
type RuleHandler interface {
	// Handle a single discovery rule
	Handle(cfg interface{}) error
	// Delete the rule and discovered targets
	Delete()
}
