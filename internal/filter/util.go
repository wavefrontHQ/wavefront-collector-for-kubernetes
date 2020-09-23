// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"fmt"
	"strings"
)

func FromQuery(vals map[string][]string) Config {
	if len(vals) == 0 {
		return Config{}
	}

	// this is legacy code retained for backwards compat when using CLI flags (instead of config file)
	// newer terminology (allow/deny) has not been back ported for now. This function and associated calls
	// can just be deleted once we choose to stop supporting the older CLI flags method for good.

	metricWhitelist := vals[MetricWhitelist]
	metricBlacklist := vals[MetricBlacklist]
	metricTagWhitelist := parseFilters(vals[MetricTagWhitelist])
	metricTagBlacklist := parseFilters(vals[MetricTagBlacklist])
	tagInclude := vals[TagInclude]
	tagExclude := vals[TagExclude]

	if len(metricWhitelist) == 0 && len(metricBlacklist) == 0 && len(metricTagWhitelist) == 0 &&
		len(metricTagBlacklist) == 0 && len(tagInclude) == 0 && len(tagExclude) == 0 {
		return Config{}
	}

	return Config{
		MetricWhitelist:    metricWhitelist,
		MetricBlacklist:    metricBlacklist,
		MetricTagWhitelist: metricTagWhitelist,
		MetricTagBlacklist: metricTagBlacklist,
		TagInclude:         tagInclude,
		TagExclude:         tagExclude,
	}
}

func FromConfig(cfg Config) Filter {
	if cfg.Empty() {
		return nil
	}

	metricAllowList := cfg.MetricWhitelist
	if len(cfg.MetricAllowList) > 0 {
		metricAllowList = cfg.MetricAllowList
	}
	metricDenyList := cfg.MetricBlacklist
	if len(cfg.MetricDenyList) > 0 {
		metricDenyList = cfg.MetricDenyList
	}
	metricTagAllowList := cfg.MetricTagWhitelist
	if len(cfg.MetricTagAllowList) > 0 {
		metricTagAllowList = cfg.MetricTagAllowList
	}
	metricTagDenyList := cfg.MetricTagBlacklist
	if len(cfg.MetricTagDenyList) > 0 {
		metricTagDenyList = cfg.MetricTagDenyList
	}
	tagInclude := cfg.TagInclude
	tagExclude := cfg.TagExclude

	if len(metricAllowList) == 0 && len(metricDenyList) == 0 && len(metricTagAllowList) == 0 &&
		len(metricTagDenyList) == 0 && len(tagInclude) == 0 && len(tagExclude) == 0 {
		return nil
	}

	return NewGlobFilter(Config{
		MetricAllowList:    metricAllowList,
		MetricDenyList:     metricDenyList,
		MetricTagAllowList: metricTagAllowList,
		MetricTagDenyList:  metricTagDenyList,
		TagInclude:         tagInclude,
		TagExclude:         tagExclude,
	})
}

func parseFilters(slice []string) map[string][]string {
	if len(slice) == 0 {
		return nil
	}
	out := make(map[string][]string)

	// each string in the slice is of the form: "tagK:[glob1, glob2, ...]"
	for _, tag := range slice {
		s := strings.Split(tag, ":")
		if len(s) == 2 {
			k, v := s[0], s[1]
			patterns, err := parseValue(v)
			if err != nil {
				fmt.Print(err)
			} else {
				out[k] = patterns
			}
		}
	}
	return out
}

// Gets a string slice from a string of the form "[foo*, bar*, ...]"
func parseValue(val string) ([]string, error) {
	if !strings.HasPrefix(val, "[") || !strings.HasSuffix(val, "]") {
		return nil, fmt.Errorf("invalid metric tag filter: %s", val)
	}
	tagValue := val[1 : len(val)-1]
	return strings.Split(tagValue, ","), nil
}
