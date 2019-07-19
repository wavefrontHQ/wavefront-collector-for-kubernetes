// Package stats provides internal metrics on the health of the Wavefront collector
package stats

import (
	"net/url"
	"sync"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	. "github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"

	metrics "github.com/rcrowley/go-metrics"
)

var doOnce sync.Once

type statsProvider struct {
	DefaultMetricsSourceProvider
	sources []MetricsSource
}

func (h *statsProvider) GetMetricsSources() []MetricsSource {
	return h.sources
}

func (h *statsProvider) Name() string {
	return "internal_stats_provider"
}

func NewInternalStatsProvider(uri *url.URL) (MetricsSourceProvider, error) {
	vals := uri.Query()
	prefix := flags.DecodeDefaultValue(vals, "prefix", "kubernetes.")
	tags := flags.DecodeTags(vals)
	filters := filter.FromQuery(vals)

	src, err := newInternalMetricsSource(prefix, tags, filters)
	if err != nil {
		return nil, err
	}
	sources := make([]MetricsSource, 1)
	sources[0] = src

	doOnce.Do(func() { // Temporal solution for https://github.com/rcrowley/go-metrics/issues/252
		metrics.RegisterRuntimeMemStats(metrics.DefaultRegistry)
	})

	return &statsProvider{
		sources: sources,
	}, nil
}
