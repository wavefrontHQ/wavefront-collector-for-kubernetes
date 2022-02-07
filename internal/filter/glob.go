// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"strings"

	"github.com/gobwas/glob"
)

type Filter interface {
	Match(name string, tags map[string]string) bool
	UsesTags() bool
}

type globFilter struct {
	metricAllowList    glob.Glob
	metricDenyList     glob.Glob
	metricTagAllowList map[string]glob.Glob
	metricTagDenyList  map[string]glob.Glob
	tagInclude         glob.Glob
	tagExclude         glob.Glob
}

func NewGlobFilter(cfg Config) Filter {
	return &globFilter{
		metricAllowList:    Compile(cfg.MetricAllowList),
		metricDenyList:     Compile(cfg.MetricDenyList),
		metricTagAllowList: MultiCompile(cfg.MetricTagAllowList),
		metricTagDenyList:  MultiCompile(cfg.MetricTagDenyList),
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

func MultiSetCompile(filters []map[string][]string) []map[string]glob.Glob {
	if len(filters) == 0 {
		return nil
	}
	globs := make([]map[string]glob.Glob, len(filters))
	for i, f := range filters {
		globs[i] = MultiCompile(f)
	}
	return globs
}

func (gf *globFilter) UsesTags() bool {
	return gf.metricTagAllowList != nil || gf.metricTagDenyList != nil ||
		gf.tagExclude != nil || gf.tagInclude != nil
}

func (gf *globFilter) Match(name string, tags map[string]string) bool {
	if gf.metricAllowList != nil && !gf.metricAllowList.Match(name) {
		return false
	}
	if gf.metricDenyList != nil && gf.metricDenyList.Match(name) {
		return false
	}

	if gf.metricTagAllowList != nil && !MatchesTags(gf.metricTagAllowList, tags) {
		return false
	}
	if gf.metricTagDenyList != nil && MatchesTags(gf.metricTagDenyList, tags) {
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

func MatchesAllTags(matchers map[string]glob.Glob, tags map[string]string) bool {
	for k, matcher := range matchers {
		if val, ok := tags[k]; ok {
			if !matcher.Match(val) {
				return false
			}
		} else {
			return false
		}
	}
	return true
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
