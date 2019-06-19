package integrations

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/integrations/telegraf"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/integrations/telegraf/memcached"
	"strings"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/integrations/telegraf/redis"
)

type discoverer struct {
	delegates map[string]discovery.Discoverer
}

// Gets a new integrations services discoverer
func NewDiscoverer(handler metrics.ProviderHandler) discovery.Discoverer {
	return &discoverer{
		delegates: makeDelegates(handler),
	}
}

func makeDelegates(handler metrics.ProviderHandler) map[string]discovery.Discoverer {
	//TODO: make this configurable
	delegates := make(map[string]discovery.Discoverer)
	delegates["redis"] = makeDiscoverer(handler, redis.NewEncoder(), "redis")
	delegates["memcached"] = makeDiscoverer(handler, memcached.NewEncoder(), "memcached")
	return delegates
}

func makeDiscoverer(handler metrics.ProviderHandler, delegate discovery.Encoder, name string) discovery.Discoverer {
	return discovery.NewDiscoverer(telegraf.NewTargetHandler(handler, delegate, discovery.NewRegistry(name)))
}

func (d *discoverer) Discover(resource discovery.Resource) {
	if resource.Kind != discovery.PodType.String() || len(resource.PodSpec.Containers) == 0 {
		// only pod discovery is supported for integrations
		return
	}

	containers := resource.PodSpec.Containers
	for _, container := range containers {
		if strings.Contains(container.Image, "redis") {
			d.delegates["redis"].Discover(resource)
		}
		if strings.Contains(container.Image, "memcached") {
			d.delegates["memcached"].Discover(resource)
		}
	}
}

func (d *discoverer) Delete(resource discovery.Resource) {
	if resource.Kind != discovery.PodType.String() || len(resource.PodSpec.Containers) == 0 {
		// only pod discovery is supported for integrations
		return
	}

	containers := resource.PodSpec.Containers
	for _, container := range containers {
		if strings.Contains(container.Image, "redis") {
			d.delegates["redis"].Delete(resource)
		}
		if strings.Contains(container.Image, "memcached") {
			d.delegates["memcached"].Delete(resource)
		}
	}
}
