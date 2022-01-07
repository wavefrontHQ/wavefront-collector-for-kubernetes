// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package stats provides internal metrics on the health of the Wavefront collector
package stats

import (
	"sync"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

	gometrics "github.com/rcrowley/go-metrics"
)

var doOnce sync.Once

type statsProvider struct {
	metrics.DefaultSourceProvider
	sources []metrics.Source
}

func (h *statsProvider) GetMetricsSources() []metrics.Source {
	return h.sources
}

func (h *statsProvider) Name() string {
	return "internal_stats_provider"
}

func NewInternalStatsProvider(cfg configuration.StatsSourceConfig) (metrics.SourceProvider, error) {
	prefix := configuration.GetStringValue(cfg.Prefix, "kubernetes.")
	tags := cfg.Tags
	filters := filter.FromConfig(cfg.Filters)

	src, err := newInternalMetricsSource(prefix, tags, filters)
	if err != nil {
		return nil, err
	}
	sources := make([]metrics.Source, 1)
	sources[0] = src

	doOnce.Do(func() { // Temporal solution for https://github.com/rcrowley/go-metrics/issues/252
		gometrics.RegisterRuntimeMemStats(gometrics.DefaultRegistry)
	})

	return &statsProvider{
		sources: sources,
	}, nil
}
