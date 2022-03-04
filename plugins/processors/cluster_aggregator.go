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

// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package processors

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

func NewClusterAggregator(metricsToAggregate []string) metrics.Processor {
	return NewSumCountAggregator("cluster", []SumCountAggregateSpec{
		{
			ResourceSumMetrics:  metricsToAggregate,
			ResourceCountMetric: metrics.MetricPodCount.Name,
			ShouldAggregate:     isType(metrics.MetricSetTypeNamespace),
			ExtractGroup:        groupByCluster,
		},
		{
			ResourceSumMetrics:  []string{},
			ResourceCountMetric: metrics.MetricPodContainerCount.Name,
			ShouldAggregate:     isType(metrics.MetricSetTypeNamespace),
			ExtractGroup:        groupByCluster,
		},
	})
}

func groupByCluster(batch *metrics.Batch, _ metrics.ResourceKey, _ *metrics.Set) (metrics.ResourceKey, *metrics.Set, error) {
	clusterSet := batch.Sets[metrics.ClusterKey()]
	if clusterSet == nil {
		clusterSet = clusterMetricSet()
	}
	return metrics.ClusterKey(), clusterSet, nil
}

func clusterMetricSet() *metrics.Set {
	return &metrics.Set{
		Values: make(map[string]metrics.Value),
		Labels: map[string]string{
			metrics.LabelMetricSetType.Key: metrics.MetricSetTypeCluster,
		},
	}
}
