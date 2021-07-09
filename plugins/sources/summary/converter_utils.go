// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package summary

import (
	"sort"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
)

const (
	sysSubContainerName = "system.slice/"
)

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
