package discovery

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PrefixAnnotation = "wavefront.com/prefix"

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

type IntegrationType int

const (
	Redis IntegrationType = 1
)

type Resource struct {
	Kind string
	IP   string
	Meta metav1.ObjectMeta

	PodSpec     v1.PodSpec
	ServiceSpec v1.ServiceSpec
}

type ResourceLister interface {
	ListPods(ns string, labels map[string]string) ([]*v1.Pod, error)
	ListServices(ns string, labels map[string]string) ([]*v1.Service, error)
}

type Discoverer interface {
	Discover(resource Resource)
	Delete(resource Resource)
}

// Handles a single discovery rule
type RuleHandler interface {
	// Handle a single discovery rule
	Handle(cfg interface{}) error
	// Delete the rule and discovered targets
	Delete()
}

type Encoder interface {
	Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) string
}

// Handles discovery of targets
type TargetHandler interface {
	Handle(resource Resource, cfg interface{})
	Encoding(name string) string
	Delete(name string)
	DeleteMissing(input map[string]bool)
	Count() int
}

// Registry for tracking discovered targets
type TargetRegistry interface {
	Register(name string, handler TargetHandler)
	Unregister(name string)
	Handler(name string) TargetHandler
	Encoding(name string) string
	Count() int
}
