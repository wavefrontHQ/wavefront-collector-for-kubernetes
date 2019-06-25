package discovery

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
)

type discoverer struct {
	delegates []discovery.Discoverer
}

func newDiscoverer(delegates ...discovery.Discoverer) discovery.Discoverer {
	return &discoverer{delegates: delegates}
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
