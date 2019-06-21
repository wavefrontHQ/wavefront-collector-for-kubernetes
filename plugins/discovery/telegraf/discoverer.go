package telegraf

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	"github.com/golang/glog"
)

type delegate struct {
	filter  *resourceFilter
	handler discovery.TargetHandler
	plugin  discovery.PluginConfig
}

type discoverer struct {
	delegates map[string]*delegate
}

// Gets a new plugins discoverer
func NewDiscoverer(handler metrics.ProviderHandler, plugins []discovery.PluginConfig) discovery.Discoverer {
	return &discoverer{
		delegates: makeDelegates(handler, plugins),
	}
}

func makeDelegates(handler metrics.ProviderHandler, plugins []discovery.PluginConfig) map[string]*delegate {
	delegates := make(map[string]*delegate)
	for _, plugin := range plugins {
		filter, err := newResourceFilter(plugin)
		if err != nil {
			glog.Errorf("error parsing plugin: %s", plugin.Type)
		}
		delegates[plugin.Type] = &delegate{
			handler: newTargetHandler(handler, plugin.Type),
			filter:  filter,
			plugin:  plugin,
		}
	}
	return delegates
}

func (d *discoverer) Discover(resource discovery.Resource) {
	if resource.Kind != discovery.PodType.String() || len(resource.PodSpec.Containers) == 0 || resource.IP == "" {
		// only pod discovery is supported here
		return
	}
	for _, delegate := range d.delegates {
		if delegate.filter.matches(resource) {
			delegate.handler.Handle(resource, delegate.plugin)
			break
		}
	}
}

func (d *discoverer) Delete(resource discovery.Resource) {
	if resource.Kind != discovery.PodType.String() || len(resource.PodSpec.Containers) == 0 {
		// only pod discovery is supported here
		return
	}
	for _, delegate := range d.delegates {
		if delegate.filter.matches(resource) {
			delegate.handler.Delete(discovery.ResourceName(resource.Kind, resource.Meta))
			break
		}
	}
}
