package prometheus

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type discoverer struct {
	runtimeHandler *targetHandler
}

func NewDiscoverer(providerHandler metrics.DynamicProviderHandler) discovery.Discoverer {
	d := &discoverer{
		runtimeHandler: newTargetHandler(providerHandler),
	}
	return d
}

func (d *discoverer) Discover(ip, kind string, meta metav1.ObjectMeta) {
	d.runtimeHandler.discover(ip, kind, meta, discovery.PrometheusConfig{})
}

func (d *discoverer) Delete(kind string, meta metav1.ObjectMeta) {
	//TODO: consider delegating to the runtime handlers
	name := resourceName(kind, meta)
	d.runtimeHandler.unregister(name)
}
