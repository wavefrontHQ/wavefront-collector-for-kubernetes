package discovery

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/prometheus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/telegraf"
)

type discoverer struct {
	delegates []discovery.Discoverer
}

func newDiscoverer(handler metrics.ProviderHandler, plugins []discovery.PluginConfig) discovery.Discoverer {
	d := &discoverer{
		delegates: make([]discovery.Discoverer, 2),
	}
	d.delegates[0] = discovery.NewDiscoverer(prometheus.NewTargetHandler(handler, true))
	d.delegates[1] = telegraf.NewDiscoverer(handler, plugins)
	return d
}

func (d *discoverer) Discover(resource discovery.Resource) {
	//TODO: make async?
	for _, delegate := range d.delegates {
		delegate.Discover(resource)
	}
}

func (d *discoverer) Delete(resource discovery.Resource) {
	//TODO: make async?
	for _, delegate := range d.delegates {
		delegate.Delete(resource)
	}
}
