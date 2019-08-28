// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/prometheus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/telegraf"
)

type delegate struct {
	filter  *resourceFilter
	handler discovery.TargetHandler
	plugin  discovery.PluginConfig
}

type discoverer struct {
	wg              sync.WaitGroup
	queue           chan discovery.Resource
	runtimeHandlers []discovery.TargetHandler
	mtx             sync.RWMutex
	delegates       map[string]*delegate
}

func newDiscoverer(handler metrics.ProviderHandler, plugins []discovery.PluginConfig) discovery.Discoverer {
	d := &discoverer{
		queue:           make(chan discovery.Resource, 1000),
		runtimeHandlers: makeRuntimeHandlers(handler),
		delegates:       makeDelegates(handler, plugins),
	}
	go d.dequeue()
	return d
}

func makeRuntimeHandlers(handler metrics.ProviderHandler) []discovery.TargetHandler {
	// currently annotation based discovery is supported only for prometheus
	return []discovery.TargetHandler{
		prometheus.NewTargetHandler(true, handler),
	}
}

func makeDelegates(handler metrics.ProviderHandler, plugins []discovery.PluginConfig) map[string]*delegate {
	delegates := make(map[string]*delegate, len(plugins))
	for _, plugin := range plugins {
		delegate, err := makeDelegate(handler, plugin)
		if err != nil {
			log.Errorf("error parsing plugin: %s error: %v", plugin.Name, err)
			continue
		}
		delegates[plugin.Name] = delegate
	}
	return delegates
}

func makeDelegate(handler metrics.ProviderHandler, plugin discovery.PluginConfig) (*delegate, error) {
	filter, err := newResourceFilter(plugin)
	if err != nil {
		return nil, err
	}
	var targetHandler discovery.TargetHandler
	if strings.Contains(plugin.Type, "prometheus") {
		targetHandler = prometheus.NewTargetHandler(false, handler)
	} else if strings.Contains(plugin.Type, "telegraf") {
		targetHandler = telegraf.NewTargetHandler(plugin.Type, handler)
	} else {
		return nil, fmt.Errorf("invalid plugin type: %s", plugin.Type)
	}
	return &delegate{
		handler: targetHandler,
		filter:  filter,
		plugin:  plugin,
	}, nil
}

func (d *discoverer) enqueue(resource discovery.Resource) {
	d.wg.Add(1)
	defer d.wg.Done()
	d.queue <- resource
}

func (d *discoverer) dequeue() {
	for resource := range d.queue {
		switch resource.Status {
		case "delete":
			d.internalDelete(resource)
		default:
			d.internalDiscover(resource)
		}
	}
	log.Infof("stopping discoverer deque")
}

func (d *discoverer) Stop() {
	d.wg.Wait()
	close(d.queue)
}

func (d *discoverer) Discover(resource discovery.Resource) {
	d.enqueue(resource)
}

func (d *discoverer) Delete(resource discovery.Resource) {
	resource.Status = "delete"
	d.enqueue(resource)
}

func (d *discoverer) internalDiscover(resource discovery.Resource) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	for _, delegate := range d.delegates {
		if delegate.filter.matches(resource) {
			delegate.handler.Handle(resource, delegate.plugin)
			return
		}
	}
	// delegate to runtime handlers if no matching delegate
	for _, runtimeHandler := range d.runtimeHandlers {
		runtimeHandler.Handle(resource, nil)
	}
}

func (d *discoverer) internalDelete(resource discovery.Resource) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	name := discovery.ResourceName(resource.Kind, resource.Meta)
	for _, delegate := range d.delegates {
		if delegate.filter.matches(resource) {
			delegate.handler.Delete(name)
			return
		}
	}
	// delegate to runtime handlers if no matching delegate
	for _, runtimeHandler := range d.runtimeHandlers {
		runtimeHandler.Delete(name)
	}
}
