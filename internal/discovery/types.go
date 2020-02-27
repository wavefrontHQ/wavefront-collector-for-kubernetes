// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

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
	PodType     ResourceType = 1
	ServiceType ResourceType = 2
	NodeType    ResourceType = 3
)

func (resType ResourceType) String() string {
	switch resType {
	case PodType:
		return "pod"
	case ServiceType:
		return "service"
	case NodeType:
		return "node"
	default:
		return fmt.Sprintf("%d", int(resType))
	}
}

// Resource encapsulates metadata about a Kubernetes resource
type Resource struct {
	Kind   string
	IP     string
	Meta   metav1.ObjectMeta
	Status string

	PodSpec     v1.PodSpec
	ServiceSpec v1.ServiceSpec
}

// Discoverer discovers endpoints from resources based on rules or annotations
type Discoverer interface {
	Discover(resource Resource)
	Delete(resource Resource)
	DeleteAll()
	Stop()
}

// Encoder generates a configuration to collect data from a given resource based on the given rules
type Encoder interface {
	Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) (string, interface{}, bool)
}

// ResourceLister lists kubernetes resources based on custom criteria
type ResourceLister interface {
	ListPods(ns string, labels map[string]string) ([]*v1.Pod, error)
	ListServices(ns string, labels map[string]string) ([]*v1.Service, error)
	ListNodes() ([]*v1.Node, error)
}

// Endpoint captures the data around a specific endpoint to collect data from
type Endpoint struct {
	Name       string
	PluginType string
	Config     interface{}
}

// EndpointHandler handles the configuration of a source to collect data from discovered endpoints
type EndpointHandler interface {
	Encode(resource Resource, rule PluginConfig) (string, interface{}, bool)
	Add(ep *Endpoint)
	Delete(ep *Endpoint)
}
