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
	err := aggregateByGroup(batch, groupByNode, isAggregatablePod, &metrics.MetricPodCount, aggregator.MetricsToAggregate)
	if err != nil {
		return nil, err
	}
	err = aggregateByGroup(batch, groupByNode, isAggregatablePodContainer, &metrics.MetricPodContainerCount, []string{})
	if err != nil {
		return nil, err
	}
	return batch, err
}

func aggregateByGroup(batch *metrics.Batch, extractGroupSet func(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set, error), shouldAggregate func(*metrics.Set) bool, count *metrics.Metric, metricsToAggregate []string) error {
	for resourceKey, resourceSet := range batch.Sets {
		if !shouldAggregate(resourceSet) {
			continue
		}
        groupKey, groupSet, err := extractGroupSet(batch, resourceKey, resourceSet)
        if err != nil {
            log.Errorf(err.Error())
            continue
        }
		aggregateCount(resourceSet, groupSet, count.Name)
		if err := aggregate(resourceSet, groupSet, metricsToAggregate); err != nil {
			return err
		}
        batch.Sets[groupKey] = groupSet
	}
	return nil
}

func groupByNode(batch *metrics.Batch, resourceKey metrics.ResourceKey, resourceSet *metrics.Set) (metrics.ResourceKey, *metrics.Set, error) {
    nodeName := resourceSet.Labels[metrics.LabelNodename.Key]
    if nodeName == "" {
        return "", nil, fmt.Errorf("no node info for resource %s: %v", resourceKey, resourceSet.Labels)
    }
    nodeKey := metrics.NodeKey(nodeName)
    return nodeKey, batch.Sets[nodeKey], nil
}

func isAggregatablePod(set *metrics.Set) bool {
	return isType(metrics.MetricSetTypePod)(set) && podTakesUpResources(set)
}

func isAggregatablePodContainer(set *metrics.Set) bool {
	return isType(metrics.MetricSetTypePodContainer)(set) && podContainerTakesUpResources(set)
}

func isType(matchType string) func(*metrics.Set) bool {
    return func(set *metrics.Set) bool {
        return set.Labels[metrics.LabelMetricSetType.Key] == matchType
    }
}

func podTakesUpResources(set *metrics.Set) bool {
	labels, _ := set.FindLabels(metrics.MetricPodPhase.Name)
	return labels["phase"] != string(corev1.PodSucceeded) && labels["phase"] != string(corev1.PodFailed)
}

func podContainerTakesUpResources(set *metrics.Set) bool {
	labels, _ := set.FindLabels(metrics.MetricContainerStatus.Name)
	return labels["state"] != "terminated"
}
