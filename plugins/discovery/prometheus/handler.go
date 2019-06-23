package prometheus

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/prometheus"
)

var registry discovery.TargetRegistry

func init() {
	registry = discovery.NewRegistry("prometheus")
}

func NewTargetHandler(handler metrics.ProviderHandler, useAnnotations bool) discovery.TargetHandler {
	return discovery.NewHandler(
		discovery.ProviderInfo{
			Handler: handler,
			Factory: prometheus.NewFactory(),
			Encoder: prometheusEncoder{},
		},
		registry,
		discovery.UseAnnotations(useAnnotations),
		discovery.SetRegistrationHandler(unregister),
	)
}

func unregister(resource discovery.Resource) bool {
	return utils.Param(resource.Meta, scrapeAnnotation, "", "false") == "false"
}
