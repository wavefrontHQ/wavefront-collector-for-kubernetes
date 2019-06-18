package discovery

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/integrations"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/discovery/prometheus"
)

type discoverer struct {
	discs []discovery.Discoverer
}

func NewDiscoverer(handler metrics.ProviderHandler) discovery.Discoverer {
	d := &discoverer{}
	d.discs = make([]discovery.Discoverer, 2)
	d.discs[0] = discovery.NewDiscoverer(prometheus.NewTargetHandler(handler, true))
	d.discs[1] = integrations.NewDiscoverer(handler)
	return d
}

func (d *discoverer) Discover(resource discovery.Resource) {
	//TODO: make async?
	for _, disc := range d.discs {
		disc.Discover(resource)
	}
}

func (d *discoverer) Delete(resource discovery.Resource) {
	//TODO: make async?
	for _, disc := range d.discs {
		disc.Delete(resource)
	}
}
