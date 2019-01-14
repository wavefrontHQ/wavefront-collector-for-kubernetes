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
	d.runtimeHandler.useAnnotations = true
	return d
}

func (d *discoverer) Discover(ip, kind string, meta metav1.ObjectMeta) {
	d.runtimeHandler.discover(ip, kind, meta, discovery.PrometheusConfig{})
}

func (d *discoverer) Delete(kind string, meta metav1.ObjectMeta) {
	name := resourceName(kind, meta)
	d.runtimeHandler.unregister(name)
}
