package prometheus

import (
	"net/url"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

type factory struct{}

// Returns a new prometheus provider factory
func NewFactory() metrics.ProviderFactory {
	return factory{}
}

func (p factory) Build(uri *url.URL) (metrics.MetricsSourceProvider, error) {
	return NewPrometheusProvider(uri)
}

func (p factory) Name() string {
	return providerName
}
