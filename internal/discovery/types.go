package discovery

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PrefixAnnotation = "wavefront.com/prefix"
	LabelsAnnotation = "wavefront.com/includeLabels"
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
	default:
		return fmt.Sprintf("%d", int(resType))
	}
}

type IntegrationType int

const (
	Redis IntegrationType = 1
)

type Resource struct {
	Kind   string
	IP     string
	Meta   metav1.ObjectMeta
	Status string

	PodSpec     v1.PodSpec
	ServiceSpec v1.ServiceSpec
}

type Discoverer interface {
	Discover(resource Resource)
	Delete(resource Resource)
	Stop()
}

// Handles the loading of discovery rules
type RuleHandler interface {
	// Handles all the discovery rules
	HandleAll(cfg []PluginConfig) error
	// Handle a single discovery rule
	Handle(cfg PluginConfig) error
	// Deletes all the rules
	DeleteAll()
	// Delete the rule and discovered targets
	Delete(name string)
	// Count of currently loaded rules
	Count() int
}

type Encoder interface {
	Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) (interface{}, bool)
}

// Handles discovery of targets
type TargetHandler interface {
	Handle(resource Resource, cfg interface{})
	Encoding(name string) interface{}
	Delete(name string)
	DeleteMissing(input map[string]bool)
	Count() int
}

// Registry for tracking discovered targets
type TargetRegistry interface {
	Register(name string, handler TargetHandler)
	Unregister(name string)
	Handler(name string) TargetHandler
	Encoding(name string) interface{}
	Count() int
}
