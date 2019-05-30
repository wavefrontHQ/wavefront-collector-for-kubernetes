// Based on https://github.com/kubernetes-retired/heapster/blob/master/metrics/sources/factory.go
// Diff against master for changes to the original code.

// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sources

import (
	"fmt"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/flags"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/prometheus"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/stats"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/summary"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/systemd"
	"github.com/wavefronthq/wavefront-kubernetes-collector/plugins/sources/telegraf"

	"github.com/golang/glog"
)

type SourceFactory struct {
}

func (sf *SourceFactory) Build(uri flags.Uri) (metrics.MetricsSourceProvider, error) {
	switch uri.Key {
	case "kubernetes.summary_api":
		provider, err := summary.NewSummaryProvider(&uri.Val)
		return provider, err
	case "prometheus":
		provider, err := prometheus.NewPrometheusProvider(&uri.Val)
		return provider, err
	case "telegraf":
		provider, err := telegraf.NewProvider(&uri.Val)
		return provider, err
	case "systemd":
		provider, err := systemd.NewProvider(&uri.Val)
		return provider, err
	default:
		return nil, fmt.Errorf("source not recognized: %s", uri.Key)
	}
}

func (sf *SourceFactory) BuildAll(uris flags.Uris, statsPrefix string) []metrics.MetricsSourceProvider {
	result := make([]metrics.MetricsSourceProvider, 0, len(uris))
	for _, uri := range uris {
		source, err := sf.Build(uri)
		if err != nil {
			glog.Errorf("Failed to create %v source: %v", uri, err)
			continue
		}
		result = append(result, source)
	}
	if len([]flags.Uri(uris)) != 0 && len(result) == 0 {
		glog.Fatal("No available source to use")
	}
	provider, _ := stats.NewInternalStatsProvider(statsPrefix)
	result = append(result, provider)
	return result
}

func NewSourceFactory() *SourceFactory {
	return &SourceFactory{}
}
