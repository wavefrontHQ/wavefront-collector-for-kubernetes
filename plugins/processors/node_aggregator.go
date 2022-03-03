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
	corev1 "k8s.io/api/core/v1"
)

// NodeAggregator aggregates MetricsToAggregate for pods by nodes. It produces by-node counts for pods and pod containers.
type NodeAggregator struct {
	MetricsToAggregate []string
}

func (aggregator *NodeAggregator) Name() string {
	return "node_aggregator"
}

func (aggregator *NodeAggregator) Process(batch *metrics.Batch) (*metrics.Batch, error) {
	err := aggregateByNode(batch, isAggregatablePod, &metrics.MetricPodCount, aggregator.MetricsToAggregate)
	if err != nil {
		return nil, err
	}
	err = aggregateByNode(batch, isAggregatablePodContainer, &metrics.MetricPodContainerCount, []string{})
	if err != nil {
		return nil, err
	}
	return batch, err
}

func aggregateByNode(batch *metrics.Batch, shouldAggregate func(*metrics.Set) bool, count *metrics.Metric, metricsToAggregate []string) error {
	for resourceKey, resourceSet := range batch.Sets {
		if !shouldAggregate(resourceSet) {
			continue
		}
		nodeName, _ := resourceSet.Labels[metrics.LabelNodename.Key]
		if nodeName == "" {
			log.Errorf("No node info for resource %s: %v", resourceKey, resourceSet.Labels)
			continue
		}
		nodeKey := metrics.NodeKey(nodeName)
		node := batch.Sets[nodeKey]
		if node == nil {
			log.Infof("No metric for node %s, cannot perform node level aggregation.", nodeKey)
			continue
		}
		aggregateCount(resourceSet, node, count.Name)
		if err := aggregate(resourceSet, node, metricsToAggregate); err != nil {
			return err
		}
	}
	return nil
}

func isAggregatablePod(set *metrics.Set) bool {
	return isType(set, metrics.MetricSetTypePod) && podTakesUpResources(set)
}

func isAggregatablePodContainer(set *metrics.Set) bool {
	return isType(set, metrics.MetricSetTypePodContainer) && podContainerTakesUpResources(set)
}

func isType(set *metrics.Set, matchType string) bool {
	return set.Labels[metrics.LabelMetricSetType.Key] == matchType
}

func podTakesUpResources(set *metrics.Set) bool {
	labels, _ := set.FindLabels(metrics.MetricPodPhase.Name)
	return labels["phase"] != string(corev1.PodSucceeded) && labels["phase"] != string(corev1.PodFailed)
}

func podContainerTakesUpResources(set *metrics.Set) bool {
	labels, _ := set.FindLabels(metrics.MetricContainerStatus.Name)
	return labels["state"] != "terminated"
}
