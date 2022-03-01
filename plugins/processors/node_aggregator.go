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

// Does not add any nodes.
type NodeAggregator struct {
	MetricsToAggregate []string
}

func (aggregator *NodeAggregator) Name() string {
	return "node_aggregator"
}

func (aggregator *NodeAggregator) Process(batch *metrics.Batch) (*metrics.Batch, error) {
	for key, metricSet := range batch.Sets {
		metricSetType, found := metricSet.Labels[metrics.LabelMetricSetType.Key]
		if !found || (metricSetType != metrics.MetricSetTypePod && metricSetType != metrics.MetricSetTypePodContainer) {
			continue
		}

		// Aggregating pods
		nodeName, found := metricSet.Labels[metrics.LabelNodename.Key]
		if nodeName == "" {
			log.Debugf("Skipping pod %s: no node info", key)
			continue
		}
		if !found {
			log.Errorf("No node info in pod %s: %v", key, metricSet.Labels)
			continue
		}
		nodeKey := metrics.NodeKey(nodeName)
		node, found := batch.Sets[nodeKey]
		if !found {
			log.Infof("No metric for node %s, cannot perform node level aggregation.", nodeKey)
		} else {
			if metricSetType == metrics.MetricSetTypePodContainer {
				// aggregate container counts and continue to top of the loop
				aggregateCount(metricSet, node, metrics.MetricPodContainerCount.Name)
				continue
			} else {
				// aggregate pod counts
				aggregateCount(metricSet, node, metrics.MetricPodCount.Name)
			}

			if err := aggregate(metricSet, node, aggregator.MetricsToAggregate); err != nil {
				return nil, err
			}
		}
	}
	return batch, nil
}
