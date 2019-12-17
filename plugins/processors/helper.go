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

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

func aggregate(src, dst *metrics.MetricSet, metricsToAggregate []string) error {
	for _, metricName := range metricsToAggregate {
		metricValue, found := src.MetricValues[metricName]
		if !found {
			continue
		}
		aggregatedValue, found := dst.MetricValues[metricName]
		if found {
			if aggregatedValue.ValueType != metricValue.ValueType {
				return fmt.Errorf("aggregator: type not supported in %s", metricName)
			}

			if aggregatedValue.ValueType == metrics.ValueInt64 {
				aggregatedValue.IntValue += metricValue.IntValue
			} else if aggregatedValue.ValueType == metrics.ValueFloat {
				aggregatedValue.FloatValue += metricValue.FloatValue
			} else {
				return fmt.Errorf("aggregator: type not supported in %s", metricName)
			}
		} else {
			aggregatedValue = metricValue
		}
		dst.MetricValues[metricName] = aggregatedValue
	}
	return nil
}

// aggregates the count of pods or containers by node, namespace and cluster.
// If the source already has aggregated counts (by namespace for example), they are used to get the counts per cluster.
// If the source does not have any counts, we increment the dest count by 1 assuming this method is invoked once per pod/container.
func aggregateCount(src, dst *metrics.MetricSet, metricName string) {
	srcCount := int64(0)
	if count, found := src.MetricValues[metricName]; found {
		srcCount += count.IntValue
	} else {
		srcCount = 1
	}

	dstCount, found := dst.MetricValues[metricName]
	if found {
		dstCount.IntValue += srcCount
	} else {
		dstCount = metrics.MetricValue{
			IntValue: srcCount,
		}
	}
	dst.MetricValues[metricName] = dstCount
}
