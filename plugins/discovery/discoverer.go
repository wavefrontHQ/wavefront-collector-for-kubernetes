package discovery

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/prometheus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/telegraf"
	"strings"
	"sync"
)

type delegate struct {
	filter  *resourceFilter
	handler discovery.TargetHandler
	plugin  discovery.PluginConfig
}

type discoverer struct {
	runtimeHandlers []discovery.TargetHandler
	mtx             sync.RWMutex
	delegates       map[string]*delegate
}

func newDiscoverer(handler metrics.ProviderHandler, plugins []discovery.PluginConfig) discovery.Discoverer {
	return &discoverer{
		runtimeHandlers: makeRuntimeHandlers(handler),
		delegates:       makeDelegates(handler, plugins),
	}
}

func makeRuntimeHandlers(handler metrics.ProviderHandler) []discovery.TargetHandler {
	// currently annotation based discovery is supported only for prometheus
	return []discovery.TargetHandler{
		prometheus.NewTargetHandler(handler, true),
	}
}

func makeDelegates(handler metrics.ProviderHandler, plugins []discovery.PluginConfig) map[string]*delegate {
	delegates := make(map[string]*delegate, len(plugins))
	for _, plugin := range plugins {
		delegate, err := makeDelegate(handler, plugin)
		if err != nil {
			glog.Errorf("error parsing plugin: %s error: %v", plugin.Name, err)
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
		targetHandler = prometheus.NewTargetHandler(handler, false)
	} else if strings.Contains(plugin.Type, "telegraf") {
		targetHandler = telegraf.NewTargetHandler(handler, plugin.Type)
	} else {
		return nil, fmt.Errorf("invalid plugin type: %s", plugin.Type)
	}
	return &delegate{
		handler: targetHandler,
		filter:  filter,
		plugin:  plugin,
	}, nil
}

func (d *discoverer) Discover(resource discovery.Resource) {
	//TODO: make async?
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

func (d *discoverer) Delete(resource discovery.Resource) {
	//TODO: make async?
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
