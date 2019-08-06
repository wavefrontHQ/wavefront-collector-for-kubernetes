// Package stats provides internal metrics on the health of the Wavefront collector
package stats

import (
	"sync"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	gometrics "github.com/rcrowley/go-metrics"
)

var doOnce sync.Once

type statsProvider struct {
	metrics.DefaultMetricsSourceProvider
	sources []metrics.MetricsSource
}

func (h *statsProvider) GetMetricsSources() []metrics.MetricsSource {
	return h.sources
}

func (h *statsProvider) Name() string {
	return "internal_stats_provider"
}

func NewInternalStatsProvider(cfg configuration.StatsSourceConfig) (metrics.MetricsSourceProvider, error) {
	prefix := configuration.GetStringValue(cfg.Prefix, "kubernetes.")
	tags := cfg.Tags
	filters := filter.FromConfig(cfg.Filters)

	src, err := newInternalMetricsSource(prefix, tags, filters)
	if err != nil {
		return nil, err
	}
	sources := make([]metrics.MetricsSource, 1)
	sources[0] = src

	doOnce.Do(func() { // Temporal solution for https://github.com/rcrowley/go-metrics/issues/252
		gometrics.RegisterRuntimeMemStats(gometrics.DefaultRegistry)
	})

	return &statsProvider{
		sources: sources,
	}, nil
}
