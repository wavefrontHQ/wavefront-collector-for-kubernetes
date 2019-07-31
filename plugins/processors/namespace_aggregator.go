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

package processors

import (
	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

type NamespaceAggregator struct {
	MetricsToAggregate []string
}

func (aggregator *NamespaceAggregator) Name() string {
	return "namespace_aggregator"
}

func (aggregator *NamespaceAggregator) Process(batch *metrics.DataBatch) (*metrics.DataBatch, error) {
	namespaces := make(map[string]*metrics.MetricSet)
	for key, metricSet := range batch.MetricSets {
		if metricSetType, found := metricSet.Labels[metrics.LabelMetricSetType.Key]; !found || metricSetType != metrics.MetricSetTypePod {
			continue
		}

		namespaceName, found := metricSet.Labels[metrics.LabelNamespaceName.Key]
		if !found {
			log.Errorf("No namespace info in pod %s: %v", key, metricSet.Labels)
			continue
		}

		namespaceKey := metrics.NamespaceKey(namespaceName)
		namespace, found := namespaces[namespaceKey]
		if !found {
			if nsFromBatch, found := batch.MetricSets[namespaceKey]; found {
				namespace = nsFromBatch
			} else {
				namespace = namespaceMetricSet(namespaceName, metricSet.Labels[metrics.LabelPodNamespaceUID.Key])
				namespaces[namespaceKey] = namespace
			}
		}

		if err := aggregate(metricSet, namespace, aggregator.MetricsToAggregate); err != nil {
			return nil, err
		}

	}
	for key, val := range namespaces {
		batch.MetricSets[key] = val
	}
	return batch, nil
}

func namespaceMetricSet(namespaceName, uid string) *metrics.MetricSet {
	return &metrics.MetricSet{
		MetricValues: make(map[string]metrics.MetricValue),
		Labels: map[string]string{
			metrics.LabelMetricSetType.Key:   metrics.MetricSetTypeNamespace,
			metrics.LabelNamespaceName.Key:   namespaceName,
			metrics.LabelPodNamespaceUID.Key: uid,
		},
	}
}
