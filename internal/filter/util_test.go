// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromConfig(t *testing.T) {
	f := FromConfig(Config{})
	assert.Equal(t, nil, f)

	// verify previous fields still work for backwards compat
	f = FromConfig(Config{
		MetricWhitelist:    []string{"foo*", "bar*"},
		MetricTagBlacklist: map[string][]string{"env": {"dev*", "test*"}},
	})

	gf, ok := f.(*globFilter)
	assert.True(t, ok)
	assert.NotNil(t, gf.metricAllowList)
	assert.NotNil(t, gf.metricTagDenyList)
	assert.Nil(t, gf.metricDenyList)
	assert.Nil(t, gf.metricTagAllowList)
	assert.Nil(t, gf.tagInclude)
	assert.Nil(t, gf.tagExclude)

	// verify new fields work
	f = FromConfig(Config{
		MetricAllowList:    []string{"foo*", "bar*"},
		MetricDenyList:     []string{"foo*", "bar*"},
		MetricTagAllowList: map[string][]string{"env": {"dev*", "test*"}},
		MetricTagDenyList:  map[string][]string{"env": {"dev*", "test*"}},
	})

	gf, ok = f.(*globFilter)
	assert.True(t, ok)
	assert.NotNil(t, gf.metricAllowList)
	assert.NotNil(t, gf.metricDenyList)
	assert.NotNil(t, gf.metricTagAllowList)
	assert.NotNil(t, gf.metricTagDenyList)
	assert.Nil(t, gf.tagInclude)
	assert.Nil(t, gf.tagExclude)
}

func TestFromQuery(t *testing.T) {
	vals := make(map[string][]string)
	if !FromQuery(vals).Empty() {
		t.Errorf("error creating filter")
	}
	vals[MetricWhitelist] = []string{"foo*", "bar*"}
	vals[MetricBlacklist] = []string{"drop*", "etc*"}
	vals[MetricTagWhitelist] = []string{"env:[foo*,bar*]", "type:[bar*]"}
	vals[MetricTagBlacklist] = []string{"env:[drop*,etc*]"}
	vals[TagInclude] = []string{"key1*", "key2*"}
	vals[TagExclude] = []string{"key3*", "key4*"}

	cfg := FromQuery(vals)
	if cfg.Empty() {
		t.Errorf("error creating filter")
	}

	f := FromConfig(cfg)

	gf, ok := f.(*globFilter)
	if !ok {
		t.Errorf("error creating filter")
	}

	if gf.metricAllowList == nil || gf.metricDenyList == nil || gf.metricTagAllowList == nil ||
		gf.metricTagDenyList == nil || gf.tagInclude == nil || gf.tagExclude == nil {
		t.Errorf("error creating filter")
	}
}

func TestParseValue(t *testing.T) {
	slice, err := parseValue("[foo*,bar*]")
	if err != nil {
		t.Errorf("error parsing value: %q", err)
	}
	expected := []string{"foo*", "bar*"}
	if !equalSlice(slice, expected) {
		t.Errorf("error parsing value, expected:%s actual:%s", expected, slice)
	}
}

func TestParseFilters(t *testing.T) {
	actual := parseFilters([]string{"env:[dev*,staging*]", "cluster:[*dev*,*staging*]"})
	expected := map[string][]string{
		"env":     {"dev*", "staging*"},
		"cluster": {"*dev*", "*staging*"},
	}
	if !equalMap(actual, expected) {
		t.Errorf("error parsing filters, expected: %q actual: %q", expected, actual)
	}
}

func equalMap(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if _, ok := b[k]; !ok {
			return false
		}
		if !equalSlice(v, b[k]) {
			return false
		}
	}
	return true
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
