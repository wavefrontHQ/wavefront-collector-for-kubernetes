// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kstate

import "github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"

func buildTags(key, name, ns string, srcTags map[string]string) map[string]string {
	tags := make(map[string]string, len(srcTags)+2)
	tags[key] = name
	tags["namespace_name"] = ns
	for k, v := range srcTags {
		tags[k] = v
	}
	return tags
}

func metricPoint(prefix, name string, value float64, ts int64, source string, tags map[string]string) *metrics.MetricPoint {
	return &metrics.MetricPoint{
		Metric:    prefix + name,
		Value:     value,
		Timestamp: ts,
		Source:    source,
		Tags:      tags,
	}
}

func floatVal(i *int32, f float64) float64 {
	if i != nil {
		return float64(*i)
	}
	return f
}
