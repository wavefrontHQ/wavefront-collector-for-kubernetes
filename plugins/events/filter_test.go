// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package events

import (
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
)

func TestAllowList(t *testing.T) {
	// test the previous field for backwards compat
	ef := newEventFilter(configuration.EventsFilter{
		TagWhitelist: map[string][]string{"foo": {"bar"}},
	})
	if ef.matches(map[string]string{"k": "v", "foo": "bard"}) {
		t.Errorf("error matching event tags")
	}
	if !ef.matches(map[string]string{"foo": "bar"}) {
		t.Errorf("error matching event tags")
	}

	ef = newEventFilter(configuration.EventsFilter{
		TagAllowList: map[string][]string{"foo": {"bar"}},
	})
	if ef.matches(map[string]string{"k": "v", "foo": "bard"}) {
		t.Errorf("error matching event tags")
	}
	if !ef.matches(map[string]string{"foo": "bar"}) {
		t.Errorf("error matching event tags")
	}
}

func TestDenyList(t *testing.T) {
	// test the previous field for backwards compat
	ef := newEventFilter(configuration.EventsFilter{
		TagBlacklist: map[string][]string{"foo": {"bar"}},
	})
	if !ef.matches(map[string]string{"k": "v", "foo": "bard"}) {
		t.Errorf("error matching event tags")
	}
	if ef.matches(map[string]string{"foo": "bar"}) {
		t.Errorf("error matching event tags")
	}

	ef = newEventFilter(configuration.EventsFilter{
		TagDenyList: map[string][]string{"foo": {"bar"}},
	})
	if !ef.matches(map[string]string{"k": "v", "foo": "bard"}) {
		t.Errorf("error matching event tags")
	}
	if ef.matches(map[string]string{"foo": "bar"}) {
		t.Errorf("error matching event tags")
	}
}

func TestAllowListSets(t *testing.T) {
	// test previous field for backwards compat
	ef := newEventFilter(configuration.EventsFilter{
		TagWhitelistSets: []map[string][]string{
			{
				"foo":  {"bar"},
				"food": {"bard"},
			},
		},
	})
	if ef.matches(map[string]string{"foo": "bar"}) {
		t.Errorf("error matching event tags")
	}
	if !ef.matches(map[string]string{"foo": "bar", "food": "bard"}) {
		t.Errorf("error matching event tags")
	}

	ef = newEventFilter(configuration.EventsFilter{
		TagAllowListSets: []map[string][]string{
			{
				"foo":  {"bar"},
				"food": {"bard"},
			},
		},
	})
	if ef.matches(map[string]string{"foo": "bar"}) {
		t.Errorf("error matching event tags")
	}
	if !ef.matches(map[string]string{"foo": "bar", "food": "bard"}) {
		t.Errorf("error matching event tags")
	}
}

func TestDenyListSets(t *testing.T) {
	// test previous field for backwards compat
	ef := newEventFilter(configuration.EventsFilter{
		TagBlacklistSets: []map[string][]string{
			{
				"foo":  {"bar"},
				"food": {"bard"},
			},
		},
	})
	if !ef.matches(map[string]string{"foo": "bar"}) {
		t.Errorf("error matching event tags")
	}
	if ef.matches(map[string]string{"foo": "bar", "food": "bard"}) {
		t.Errorf("error matching event tags")
	}

	ef = newEventFilter(configuration.EventsFilter{
		TagDenyListSets: []map[string][]string{
			{
				"foo":  {"bar"},
				"food": {"bard"},
			},
		},
	})
	if !ef.matches(map[string]string{"foo": "bar"}) {
		t.Errorf("error matching event tags")
	}
	if ef.matches(map[string]string{"foo": "bar", "food": "bard"}) {
		t.Errorf("error matching event tags")
	}
}
