// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"strings"
	"testing"

	"github.com/gobwas/glob"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/metrics"
)

func TestMatchesTag(t *testing.T) {
	matcher := compileGlob([]string{"foo"}, t)
	if !matchesTag(matcher, map[string]string{"foo": "bar", "key1": "val1"}) {
		t.Errorf("error matching tag")
	}
	if matchesTag(matcher, map[string]string{"foobar": "bar", "key1": "val1"}) {
		t.Errorf("error matching tag")
	}

	matcher = compileGlob([]string{"foo*"}, t)
	if !matchesTag(matcher, map[string]string{"foobar": "bar", "key1": "val1"}) {
		t.Errorf("error matching tag")
	}
}

func TestMatchesTags(t *testing.T) {
	matchers := MultiCompile(map[string][]string{
		"env":  {"?rod1*", "prod2*"},
		"type": {"pod", "service"},
		"node": {"10.2.*", "10.3.*"},
	})

	if MatchesTags(matchers, map[string]string{"env": "prod"}) {
		t.Errorf("error matching tags")
	}
	if !MatchesTags(matchers, map[string]string{"env": "prod234"}) {
		t.Errorf("error matching tags")
	}
	if !MatchesTags(matchers, map[string]string{"env": "prod134"}) {
		t.Errorf("error matching tags")
	}
	if !MatchesTags(matchers, map[string]string{"type": "service"}) {
		t.Errorf("error matching tags")
	}
	if !MatchesTags(matchers, map[string]string{"node": "10.2.45.2"}) {
		t.Errorf("error matching tags")
	}
}

func TestDeleteTags(t *testing.T) {
	// test tagIncludes: only matching tags should remain in the map
	matcher := compileGlob([]string{"foo"}, t)
	tags := map[string]string{"foo": "bar", "key1": "val1", "key2": "val2", "foobar": "bar"}
	deleteTags(matcher, tags, true)
	if len(tags) != 1 {
		t.Errorf("error deleting tags")
	}
	if _, ok := tags["foo"]; !ok {
		t.Errorf("error deleting tags")
	}

	// test tagExcludes: excluded tags should be removed
	matcher = compileGlob([]string{"foo*"}, t)
	tags = map[string]string{"foo": "bar", "key1": "val1", "key2": "val2", "foobar": "bar"}
	deleteTags(matcher, tags, false)
	if len(tags) != 2 {
		t.Errorf("error deleting tags")
	}
	if _, ok := tags["foo"]; ok {
		t.Errorf("error deleting tags")
	}
	if _, ok := tags["foobar"]; ok {
		t.Errorf("error deleting tags")
	}
}

func TestMetricWhitelist(t *testing.T) {
	cfg := Config{
		MetricWhitelist: []string{"foo"},
	}
	f := NewGlobFilter(cfg)

	pt := point("foobar", 1.0, 0, "", nil)
	if f.Match(pt.Metric, pt.Tags) {
		t.Errorf("name pass error")
	}

	pt = point("foo", 1.0, 0, "", nil)
	if !f.Match(pt.Metric, pt.Tags) {
		t.Errorf("name pass error")
	}

	cfg.MetricWhitelist = []string{"foo*"}
	f = NewGlobFilter(cfg)
	if !f.Match(pt.Metric, pt.Tags) {
		t.Errorf("name pass error")
	}
}

func TestMetricBlacklist(t *testing.T) {
	cfg := Config{
		MetricBlacklist: []string{"foo"},
	}
	f := NewGlobFilter(cfg)
	pt := point("foobar", 1.0, 0, "", nil)
	if !f.Match(pt.Metric, pt.Tags) {
		t.Errorf("name drop error")
	}

	cfg.MetricBlacklist = []string{"foo*"}
	f = NewGlobFilter(cfg)
	if f.Match(pt.Metric, pt.Tags) {
		t.Errorf("name drop error")
	}
}

func TestMetricTagWhitelist(t *testing.T) {
	cfg := Config{
		MetricTagWhitelist: map[string][]string{
			"foo": {"va*"},
		},
	}
	f := NewGlobFilter(cfg)
	pt := point("bar", 1.0, 0, "", map[string]string{"bar": "foo"})
	if f.Match(pt.Metric, pt.Tags) {
		t.Errorf("tag pass error")
	}

	pt = point("bar", 1.0, 0, "", map[string]string{"bar": "foo", "foo": "val"})
	if !f.Match(pt.Metric, pt.Tags) {
		t.Errorf("tag pass error")
	}
}

func TestMetricTagBlacklist(t *testing.T) {
	cfg := Config{
		MetricTagBlacklist: map[string][]string{
			"foo": {"va*"},
		},
	}
	f := NewGlobFilter(cfg)
	pt := point("bar", 1.0, 0, "", map[string]string{"bar": "foo"})
	if !f.Match(pt.Metric, pt.Tags) {
		t.Errorf("tag drop error")
	}

	pt = point("bar", 1.0, 0, "", map[string]string{"bar": "foo", "foo": "val"})
	if f.Match(pt.Metric, pt.Tags) {
		t.Errorf("tag drop error")
	}
}

func TestTagInclude(t *testing.T) {
	cfg := Config{
		TagInclude: []string{"foo*"},
	}
	f := NewGlobFilter(cfg)
	pt := point("bar", 1.0, 0, "", map[string]string{"foo": "bar", "key1": "val1"})
	if !f.Match(pt.Metric, pt.Tags) {
		t.Errorf("tag include error")
	}
	if len(pt.Tags) != 1 {
		t.Errorf("tag include error")
	}
	if _, ok := pt.Tags["foo"]; !ok {
		t.Errorf("tag include error")
	}
}

func TestTagExclude(t *testing.T) {
	cfg := Config{
		TagExclude: []string{"foo*"},
	}
	f := NewGlobFilter(cfg)
	pt := point("bar", 1.0, 0, "", map[string]string{"foo": "bar", "key1": "val1"})
	if !f.Match(pt.Metric, pt.Tags) {
		t.Errorf("tag exclude error")
	}
	if len(pt.Tags) != 1 {
		t.Errorf("tag exclude error")
	}
	if _, ok := pt.Tags["foo"]; ok {
		t.Errorf("tag exclude error")
	}
}

func point(name string, value float64, ts int64, source string, tags map[string]string) *metrics.MetricPoint {
	return &metrics.MetricPoint{
		Metric:    strings.Replace(name, "_", ".", -1),
		Value:     value,
		Timestamp: ts,
		Source:    source,
		Tags:      tags,
	}
}

func compileGlob(filter []string, t *testing.T) glob.Glob {
	matcher := Compile(filter)
	if matcher == nil {
		t.Errorf("error creating matcher: %q", filter)
	}
	return matcher
}
