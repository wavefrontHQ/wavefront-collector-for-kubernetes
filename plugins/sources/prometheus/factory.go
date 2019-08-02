package prometheus

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

type factory struct{}

// Returns a new prometheus provider factory
func NewFactory() metrics.ProviderFactory {
	return factory{}
}

func (p factory) Build(cfg interface{}) (metrics.MetricsSourceProvider, error) {
	c := cfg.(configuration.PrometheusSourceConfig)
	provider, err := NewPrometheusProvider(c)
	if err == nil {
		if i, ok := provider.(metrics.ConfigurabeMetricsSourceProvider); ok {
			i.Configure(c.Collection.Interval, c.Collection.Timeout)
		}
	}
	return provider, err
}

func (p factory) Name() string {
	return providerName
}
