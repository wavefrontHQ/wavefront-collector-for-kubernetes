// Package stats provides internal metrics on the health of the Wavefront collector
package stats

import (
	"time"

	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	metrics "github.com/rcrowley/go-metrics"
)

var statsPrefix string

type internalMetricsSource struct{}

func (src *internalMetricsSource) Name() string {
	return "internal_stats_source"
}

func (src *internalMetricsSource) ScrapeMetrics() (*DataBatch, error) {
	return internalStats()
}

type statsProvider struct {
	sources []MetricsSource
}

func (h *statsProvider) GetMetricsSources() []MetricsSource {
	return h.sources
}

func (h *statsProvider) Name() string {
	return "internal_stats_provider"
}

func (h *statsProvider) CollectionInterval() time.Duration {
	return time.Duration(10 * time.Second)
}

func (h *statsProvider) TimeOut() time.Duration {
	return time.Duration(10 * time.Second)
}

func NewInternalStatsProvider(prefix string) (MetricsSourceProvider, error) {
	sources := make([]MetricsSource, 1)
	sources[0] = &internalMetricsSource{}
	metrics.RegisterRuntimeMemStats(metrics.DefaultRegistry)

	// temp workaround. remove once sink level prefix is applied to all metrics.
	statsPrefix = prefix

	return &statsProvider{
		sources: sources,
	}, nil
}
