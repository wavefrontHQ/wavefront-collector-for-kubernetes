// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package wavefront

import (
	"sort"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

func processTags(tags map[string]string) {
	for k, v := range tags {
		// ignore tags with empty values as well so the data point doesn't fail validation
		if excludeTag(k) || len(v) == 0 {
			delete(tags, k)
		}
	}
}

func excludeTag(a string) bool {
	for _, b := range excludeTagList {
		if b == a {
			return true
		}
	}
	return false
}

func sortedMetricSetKeys(m map[string]*metrics.MetricSet) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func sortedLabelKeys(m map[string]string) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func sortedMetricValueKeys(m map[string]metrics.MetricValue) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
