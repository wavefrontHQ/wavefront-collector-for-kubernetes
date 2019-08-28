// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package telegraf

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/configuration"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

type factory struct{}

// Returns a new telegraf provider factory
func NewFactory() metrics.ProviderFactory {
	return factory{}
}

func (p factory) Build(cfg interface{}) (metrics.MetricsSourceProvider, error) {
	c := cfg.(configuration.TelegrafSourceConfig)
	provider, err := NewProvider(c)
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
