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
	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

func NewNodeAggregator(metricsToAggregate []string) metrics.Processor {
	return NewSumCountAggregator("node", []SumCountAggregateSpec{
		{
			ResourceSumMetrics:  metricsToAggregate,
			ResourceCountMetric: metrics.MetricPodCount.Name,
			isPartOfGroup:       isAggregatablePod,
			Group:               nodeGroup,
		},
		{
			ResourceSumMetrics:  []string{},
			ResourceCountMetric: metrics.MetricPodContainerCount.Name,
			isPartOfGroup:       isAggregatablePodContainer,
			Group:               nodeGroup,
		},
	})
}

func nodeGroup(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set) {
	nodeName := resourceSet.Labels[metrics.LabelNodename.Key]
	if nodeName == "" {
		log.Errorf("no node info for resource %s: %v", resourceKey, resourceSet.Labels)
		return "", nil
	}
	nodeKey := metrics.NodeKey(nodeName)
	return nodeKey, batch.Sets[nodeKey]
}
