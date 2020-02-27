// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/prometheus"
)

func NewProviderInfo(handler metrics.ProviderHandler, prefix string) discovery.ProviderInfo {
	return discovery.ProviderInfo{
		Handler: handler,
		Factory: prometheus.NewFactory(),
		Encoder: newPrometheusEncoder(prefix),
	}
}
