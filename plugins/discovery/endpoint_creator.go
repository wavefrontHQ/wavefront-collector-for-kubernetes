// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	log "github.com/sirupsen/logrus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
)

type endpointCreator struct {
	delegates                  map[string]*delegate
	providers                  map[string]discovery.ProviderInfo
	disableAnnotationDiscovery bool
	annotationExcludes         []*resourceFilter
}

func (e *endpointCreator) discoverEndpointsWithRules(resource discovery.Resource) []*discovery.Endpoint {
	var eps []*discovery.Endpoint
	for _, delegate := range e.delegates {
		if delegate.filter.matches(resource) {
			if ep := e.makeEndpoint(resource, delegate.plugin); ep != nil {
				eps = append(eps, ep)
			}
		}
	}
	return eps
}

func (e *endpointCreator) discoverEndpointsWithAnnotations(resource discovery.Resource) []*discovery.Endpoint {
	for _, exclude := range e.annotationExcludes {
		if exclude.matches(resource) {
			return nil
		}
	}
	var eps []*discovery.Endpoint
	if ep := e.makeEndpoint(resource, discovery.PluginConfig{Type: "prometheus"}); ep != nil {
		eps = append(eps, ep)
	}

	return eps
}

func (e *endpointCreator) discoverEndpoints(resource discovery.Resource) []*discovery.Endpoint {
	eps := e.discoverEndpointsWithRules(resource)
	if len(eps) == 0 && !e.disableAnnotationDiscovery {
		eps = e.discoverEndpointsWithAnnotations(resource)
	}
	return eps
}

func (e *endpointCreator) makeEndpoint(resource discovery.Resource, plugin discovery.PluginConfig) *discovery.Endpoint {
	if name, cfg, ok := e.Encode(resource, plugin); ok {
		return &discovery.Endpoint{
			Name:       name,
			Config:     cfg,
			PluginType: pluginType(plugin),
		}
	}
	return nil
}

func (e *endpointCreator) Encode(resource discovery.Resource, rule discovery.PluginConfig) (string, interface{}, bool) {
	kind := resource.Kind
	ip := resource.IP
	meta := resource.Meta

	if log.IsLevelEnabled(log.DebugLevel) {
		log.WithFields(log.Fields{
			"kind":      kind,
			"name":      meta.Name,
			"namespace": meta.Namespace,
		}).Debug("handling resource")
	}

	if delegate, ok := e.providers[pluginType(rule)]; ok {
		return delegate.Encoder.Encode(ip, kind, meta, rule)
	}
	return "", nil, false
}
