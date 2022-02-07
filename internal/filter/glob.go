// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"strings"

	"github.com/gobwas/glob"
)

type Filter interface {
	MatchMetric(name string, tags map[string]string) bool
	MatchTag(tagName string) bool
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

func (gf *globFilter) MatchMetric(name string, tags map[string]string) bool {
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

	return true
}

func (gf *globFilter) MatchTag(tagName string) bool {
	matches := true
	if gf.tagInclude != nil {
		matches = matches && gf.tagInclude.Match(tagName)
	}
	if gf.tagExclude != nil {
		matches = matches && !gf.tagExclude.Match(tagName)
	}
	return matches
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
