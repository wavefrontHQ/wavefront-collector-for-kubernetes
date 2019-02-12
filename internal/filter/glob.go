package filter

import (
	"strings"

	"github.com/gobwas/glob"
)

type Filter interface {
	Match(name string, tags map[string]string) bool
}

type globFilter struct {
	namePass   glob.Glob
	nameDrop   glob.Glob
	tagPass    map[string]glob.Glob
	tagDrop    map[string]glob.Glob
	tagInclude glob.Glob
	tagExclude glob.Glob
}

func NewGlobFilter(cfg Config) Filter {
	return &globFilter{
		namePass:   compile(cfg.MetricWhitelist),
		nameDrop:   compile(cfg.MetricBlacklist),
		tagPass:    multiCompile(cfg.MetricTagWhitelist),
		tagDrop:    multiCompile(cfg.MetricTagBlacklist),
		tagInclude: compile(cfg.TagInclude),
		tagExclude: compile(cfg.TagExclude),
	}
}

func compile(filters []string) glob.Glob {
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

func multiCompile(filters map[string][]string) map[string]glob.Glob {
	if len(filters) == 0 {
		return nil
	}
	globs := make(map[string]glob.Glob, len(filters))
	for k, v := range filters {
		g := compile(v)
		if g != nil {
			globs[k] = g
		}
	}
	return globs
}

func (gf *globFilter) Match(name string, tags map[string]string) bool {
	if gf.namePass != nil && !gf.namePass.Match(name) {
		return false
	}
	if gf.nameDrop != nil && gf.nameDrop.Match(name) {
		return false
	}

	if gf.tagPass != nil && !matchesTags(gf.tagPass, tags) {
		return false
	}
	if gf.tagDrop != nil && matchesTags(gf.tagDrop, tags) {
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

func matchesTags(matchers map[string]glob.Glob, tags map[string]string) bool {
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
