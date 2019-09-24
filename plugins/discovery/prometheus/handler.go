// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery/utils"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/prometheus"
)

func NewTargetHandler(useAnnotations bool, handler metrics.ProviderHandler, prefix string) discovery.TargetHandler {
	if prefix == "" {
		prefix = "prometheus.io"
	}
	scrapeAnnotation := customAnnotation(scrapeAnnotationFormat, prefix)
	return discovery.NewHandler(
		discovery.ProviderInfo{
			Handler: handler,
			Factory: prometheus.NewFactory(),
			Encoder: newPrometheusEncoder(prefix),
		},
		discovery.NewRegistry("prometheus"),
		discovery.UseAnnotations(useAnnotations),
		discovery.SetRegistrationHandler(func(resource discovery.Resource) bool {
			return utils.Param(resource.Meta, scrapeAnnotation, "", "false") == "false"
		}),
	)
}
