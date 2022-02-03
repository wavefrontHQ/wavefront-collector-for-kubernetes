// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/wf"

	"github.com/gobwas/glob"

	"github.com/stretchr/testify/assert"
)

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

func TestMatchesAllTags(t *testing.T) {
	matchers := MultiCompile(map[string][]string{
		"env":  {"?rod1*", "prod2*"},
		"type": {"pod", "service"},
		"node": {"10.2.*", "10.3.*"},
	})

	if MatchesAllTags(matchers, map[string]string{"env": "prod134"}) {
		t.Errorf("error matching all tags")
	}

	if !MatchesAllTags(matchers, map[string]string{"env": "prod234", "type": "pod", "node": "10.2.3.4"}) {
		t.Errorf("error matching all tags")
	}
}

func TestMetricAllowList(t *testing.T) {
	cfg := Config{
		MetricAllowList: []string{"foo"},
	}
	f := NewGlobFilter(cfg)

	pt := wf.NewPoint("foobar", 1.0, 0, "", nil)
	if f.MatchMetric(pt.Metric, pt.Tags()) {
		t.Errorf("name pass error")
	}

	pt = wf.NewPoint("foo", 1.0, 0, "", nil)
	if !f.MatchMetric(pt.Metric, pt.Tags()) {
		t.Errorf("name pass error")
	}

	cfg.MetricAllowList = []string{"foo*"}
	f = NewGlobFilter(cfg)
	if !f.MatchMetric(pt.Metric, pt.Tags()) {
		t.Errorf("name pass error")
	}
}

func TestMetricDenyList(t *testing.T) {
	cfg := Config{
		MetricDenyList: []string{"foo"},
	}
	f := NewGlobFilter(cfg)
	pt := wf.NewPoint("foobar", 1.0, 0, "", nil)
	if !f.MatchMetric(pt.Metric, pt.Tags()) {
		t.Errorf("name drop error")
	}

	cfg.MetricDenyList = []string{"foo*"}
	f = NewGlobFilter(cfg)
	if f.MatchMetric(pt.Metric, pt.Tags()) {
		t.Errorf("name drop error")
	}
}

func TestMetricTagAllowList(t *testing.T) {
	cfg := Config{
		MetricTagAllowList: map[string][]string{
			"foo": {"va*"},
		},
	}
	f := NewGlobFilter(cfg)
	pt := wf.NewPoint("bar", 1.0, 0, "", map[string]string{"bar": "foo"})
	if f.MatchMetric(pt.Metric, pt.Tags()) {
		t.Errorf("tag pass error")
	}

	pt = wf.NewPoint("bar", 1.0, 0, "", map[string]string{"bar": "foo", "foo": "val"})
	if !f.MatchMetric(pt.Metric, pt.Tags()) {
		t.Errorf("tag pass error")
	}
}

func TestMetricTagDenyList(t *testing.T) {
	cfg := Config{
		MetricTagDenyList: map[string][]string{
			"foo": {"va*"},
		},
	}
	f := NewGlobFilter(cfg)
	pt := wf.NewPoint("bar", 1.0, 0, "", map[string]string{"bar": "foo"})
	if !f.MatchMetric(pt.Metric, pt.Tags()) {
		t.Errorf("tag drop error")
	}

	pt = wf.NewPoint("bar", 1.0, 0, "", map[string]string{"bar": "foo", "foo": "val"})
	if f.MatchMetric(pt.Metric, pt.Tags()) {
		t.Errorf("tag drop error")
	}
}

func TestTagInclude(t *testing.T) {
	f := NewGlobFilter(Config{
		TagInclude: []string{"foo*"},
	})

	assert.True(t, f.MatchTag("foo"))
	assert.True(t, f.MatchTag("foobar"))
	assert.False(t, f.MatchTag("barfoo"))
}

func TestTagExclude(t *testing.T) {
	f := NewGlobFilter(Config{
		TagExclude: []string{"foo*"},
	})

	assert.False(t, f.MatchTag("foo"))
	assert.False(t, f.MatchTag("foobar"))
	assert.True(t, f.MatchTag("barfoo"))
}

func compileGlob(filter []string, t *testing.T) glob.Glob {
	matcher := Compile(filter)
	if matcher == nil {
		t.Errorf("error creating matcher: %q", filter)
	}
	return matcher
}
