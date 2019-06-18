package integrations

import (
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
	delegates["redis"] = discovery.NewDiscoverer(redis.NewTargetHandler(handler))
	return delegates
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
	}
}
