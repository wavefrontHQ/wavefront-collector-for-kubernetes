package prometheus

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/prometheus"
)

func NewTargetHandler(useAnnotations bool) discovery.TargetHandler {
	return discovery.NewHandler(
		discovery.ProviderInfo{
			Factory: prometheus.NewFactory(),
			Encoder: prometheusEncoder{},
		},
		discovery.NewRegistry("prometheus"),
		discovery.UseAnnotations(useAnnotations),
		discovery.SetRegistrationHandler(unregister),
	)
}

func unregister(resource discovery.Resource) bool {
	return utils.Param(resource.Meta, scrapeAnnotation, "", "false") == "false"
}
