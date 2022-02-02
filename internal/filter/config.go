// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

const (
	MetricWhitelist    = "metricWhitelist"
	MetricBlacklist    = "metricBlacklist"
	MetricTagWhitelist = "metricTagWhitelist"
	MetricTagBlacklist = "metricTagBlacklist"
	TagInclude         = "tagInclude"
	TagExclude         = "tagExclude"
)

// Configuration for filtering metrics.
// All the filtering options are applied at the end after specified prefixes etc are applied.
type Config struct {
	// List of glob pattern strings. Only metrics with names matching this list are reported.
	MetricAllowList []string `yaml:"metricAllowList"`

	// List of glob pattern strings. Metrics with names matching this list are dropped.
	MetricDenyList []string `yaml:"metricDenyList"`

	// List of glob pattern strings. Only metrics containing tag keys matching the list will be reported.
	MetricTagAllowList map[string][]string `yaml:"metricTagAllowList"`

	// List of glob pattern strings. Metrics containing these tag keys will be dropped.
	MetricTagDenyList map[string][]string `yaml:"metricTagDenyList"`

	// List of glob pattern strings. tags with matching keys will be included. All other tags will be excluded.
	TagInclude []string `yaml:"tagInclude"`

	// List of glob pattern strings. tags with matching keys will be excluded.
	TagExclude []string `yaml:"tagExclude"`

	// Deprecated: use MetricAllowList instead
	MetricWhitelist []string `yaml:"metricWhitelist"`

	// Deprecated: use MetricDenyList instead
	MetricBlacklist []string `yaml:"metricBlacklist"`

	// Deprecated: use MetricTagAllowList instead
	MetricTagWhitelist map[string][]string `yaml:"metricTagWhitelist"`

	// Deprecated: use MetricTagDenyList instead
	MetricTagBlacklist map[string][]string `yaml:"metricTagBlacklist"`
}

func (cfg Config) Empty() bool {
	return len(cfg.MetricWhitelist) == 0 && len(cfg.MetricAllowList) == 0 &&
		len(cfg.MetricBlacklist) == 0 && len(cfg.MetricDenyList) == 0 &&
		len(cfg.MetricTagWhitelist) == 0 && len(cfg.MetricTagAllowList) == 0 &&
		len(cfg.MetricTagBlacklist) == 0 && len(cfg.MetricTagDenyList) == 0 &&
		len(cfg.TagInclude) == 0 && len(cfg.TagExclude) == 0
}
