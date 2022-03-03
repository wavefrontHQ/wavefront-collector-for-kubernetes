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
	"fmt"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

type NamespaceAggregator struct {
	MetricsToAggregate []string
}

func (aggregator *NamespaceAggregator) Name() string {
	return "namespace_aggregator"
}

func (aggregator *NamespaceAggregator) Process(batch *metrics.Batch) (*metrics.Batch, error) {
	err := aggregateByGroup(batch, groupByNamespace, isAggregatablePod, &metrics.MetricPodCount, aggregator.MetricsToAggregate)
	if err != nil {
		return nil, err
	}
	err = aggregateByGroup(batch, groupByNamespace, isAggregatablePodContainer, &metrics.MetricPodContainerCount, []string{})
	if err != nil {
		return nil, err
	}
	return batch, err
}

func groupByNamespace(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set, error) {
	namespaceName, found := resourceSet.Labels[metrics.LabelNamespaceName.Key]
	if !found {
		return "", nil, fmt.Errorf("no namespace info in pod %s: %v", resourceKey, resourceSet.Labels)
	}
	namespaceKey := metrics.NamespaceKey(namespaceName)
	namespaceSet := batch.Sets[namespaceKey]
	if namespaceSet == nil {
		namespaceSet = namespaceMetricSet(namespaceName, resourceSet.Labels[metrics.LabelPodNamespaceUID.Key])
	}
	return namespaceKey, namespaceSet, nil
}

func namespaceMetricSet(namespaceName, uid string) *metrics.Set {
	return &metrics.Set{
		Values: make(map[string]metrics.Value),
		Labels: map[string]string{
			metrics.LabelMetricSetType.Key:   metrics.MetricSetTypeNamespace,
			metrics.LabelNamespaceName.Key:   namespaceName,
			metrics.LabelPodNamespaceUID.Key: uid,
		},
	}
}
