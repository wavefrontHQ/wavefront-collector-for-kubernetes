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

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

var LabelsToPopulate = []metrics.LabelDescriptor{
	metrics.LabelPodId,
	metrics.LabelPodName,
	metrics.LabelNamespaceName,
	metrics.LabelPodNamespaceUID,
	metrics.LabelHostname,
	metrics.LabelHostID,
}

type PodAggregator struct {
	skippedMetrics map[string]struct{}
}

func (aggregator *PodAggregator) Name() string {
	return "pod_aggregator"
}

func (aggregator *PodAggregator) Process(batch *metrics.DataBatch) (*metrics.DataBatch, error) {
	newPods := make(map[string]*metrics.MetricSet)

	// If pod already has pod-level metrics, it no longer needs to aggregates its container's metrics.
	requireAggregate := make(map[string]bool)
	for key, metricSet := range batch.MetricSets {
		if metricSetType, found := metricSet.Labels[metrics.LabelMetricSetType.Key]; !found || metricSetType != metrics.MetricSetTypePodContainer {
			continue
		}

		// Aggregating containers
		podName, found := metricSet.Labels[metrics.LabelPodName.Key]
		ns, found2 := metricSet.Labels[metrics.LabelNamespaceName.Key]
		if !found || !found2 {
			log.Errorf("No namespace and/or pod info in container %s: %v", key, metricSet.Labels)
			continue
		}

		podKey := metrics.PodKey(ns, podName)
		pod, found := batch.MetricSets[podKey]
		if !found {
			pod, found = newPods[podKey]
			if !found {
				log.Infof("Pod not found adding %s", podKey)
				pod = aggregator.podMetricSet(metricSet.Labels)
				newPods[podKey] = pod
			}
		}

		for metricName, metricValue := range metricSet.MetricValues {
			if _, found := aggregator.skippedMetrics[metricName]; found {
				continue
			}

			aggregatedValue, found := pod.MetricValues[metricName]
			if !found {
				requireAggregate[podKey+metricName] = true
				aggregatedValue = metricValue
			} else {
				if requireAggregate[podKey+metricName] {
					if aggregatedValue.ValueType != metricValue.ValueType {
						log.Errorf("PodAggregator: inconsistent type in %s", metricName)
						continue
					}

					switch aggregatedValue.ValueType {
					case metrics.ValueInt64:
						aggregatedValue.IntValue += metricValue.IntValue
					case metrics.ValueFloat:
						aggregatedValue.FloatValue += metricValue.FloatValue
					default:
						return nil, fmt.Errorf("PodAggregator: type not supported in %s", metricName)
					}
				}
			}

			pod.MetricValues[metricName] = aggregatedValue
		}
	}
	for key, val := range newPods {
		batch.MetricSets[key] = val
	}
	return batch, nil
}

func (aggregator *PodAggregator) podMetricSet(labels map[string]string) *metrics.MetricSet {
	newLabels := map[string]string{
		metrics.LabelMetricSetType.Key: metrics.MetricSetTypePod,
	}
	for _, l := range LabelsToPopulate {
		if val, ok := labels[l.Key]; ok {
			newLabels[l.Key] = val
		}
	}
	return &metrics.MetricSet{
		MetricValues: make(map[string]metrics.MetricValue),
		Labels:       newLabels,
	}
}

func NewPodAggregator() *PodAggregator {
	skipped := make(map[string]struct{})
	for _, metric := range metrics.StandardMetrics {
		if metric.MetricDescriptor.Type == metrics.MetricCumulative ||
			metric.MetricDescriptor.Type == metrics.MetricDelta {
			skipped[metric.MetricDescriptor.Name] = struct{}{}
		}
	}
	return &PodAggregator{
		skippedMetrics: skipped,
	}
}
