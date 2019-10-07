// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"strings"

	"github.com/gobwas/glob"
)

type Filter interface {
	Match(name string, tags map[string]string) bool
}

type globFilter struct {
	metricWhitelist    glob.Glob
	metricBlacklist    glob.Glob
	metricTagWhitelist map[string]glob.Glob
	metricTagBlacklist map[string]glob.Glob
	tagInclude         glob.Glob
	tagExclude         glob.Glob
}

func NewGlobFilter(cfg Config) Filter {
	return &globFilter{
		metricWhitelist:    Compile(cfg.MetricWhitelist),
		metricBlacklist:    Compile(cfg.MetricBlacklist),
		metricTagWhitelist: MultiCompile(cfg.MetricTagWhitelist),
		metricTagBlacklist: MultiCompile(cfg.MetricTagBlacklist),
		tagInclude:         Compile(cfg.TagInclude),
		tagExclude:         Compile(cfg.TagExclude),
	}
}

func Compile(filters []string) glob.Glob {
	if len(filters) == 0 {
		return nil
	}
	if len(filters) == 1 {
		g, _ := glob.Compile(filters[0])
		return g
	}
	g, _ := glob.Compile("{" + strings.Join(filters, ",") + "}")
	return g
}

func MultiCompile(filters map[string][]string) map[string]glob.Glob {
	if len(filters) == 0 {
		return nil
	}
	globs := make(map[string]glob.Glob, len(filters))
	for k, v := range filters {
		g := Compile(v)
		if g != nil {
			globs[k] = g
		}
	}
	return globs
}

func (gf *globFilter) Match(name string, tags map[string]string) bool {
	if gf.metricWhitelist != nil && !gf.metricWhitelist.Match(name) {
		return false
	}
	if gf.metricBlacklist != nil && gf.metricBlacklist.Match(name) {
		return false
	}

	if gf.metricTagWhitelist != nil && !MatchesTags(gf.metricTagWhitelist, tags) {
		return false
	}
	if gf.metricTagBlacklist != nil && MatchesTags(gf.metricTagBlacklist, tags) {
		return false
	}

	if gf.tagInclude != nil {
		deleteTags(gf.tagInclude, tags, true)
	}
	if gf.tagExclude != nil {
		deleteTags(gf.tagExclude, tags, false)
	}
	return true
}

func MatchesTags(matchers map[string]glob.Glob, tags map[string]string) bool {
	for k, matcher := range matchers {
		if val, ok := tags[k]; ok {
			if matcher.Match(val) {
				return true
			}
		}
	}
	return false
}

func matchesTag(matcher glob.Glob, tags map[string]string) bool {
	for k := range tags {
		if matcher.Match(k) {
			return true
		}
	}
	return false
}

func deleteTags(matcher glob.Glob, tags map[string]string, include bool) {
	for k := range tags {
		matches := matcher.Match(k)
		if (include && !matches) || (!include && matches) {
			delete(tags, k)
		}
	}
}
