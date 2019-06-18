package prometheus

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/prometheus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type prometheusEncoder struct{}

func (e prometheusEncoder) Encode(ip, kind string, meta metav1.ObjectMeta, rule interface{}) string {
	cfg := discovery.PrometheusConfig{}
	if rule != nil {
		cfg = rule.(discovery.PrometheusConfig)
	}
	return scrapeURL(ip, kind, meta, cfg)
}

func unregister(resource discovery.Resource) bool {
	return utils.Param(resource.Meta, scrapeAnnotation, "", "false") == "false"
}
