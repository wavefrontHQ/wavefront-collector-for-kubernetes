// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package events

import (
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
)

func TestWhitelist(t *testing.T) {
	ef := newEventFilter(configuration.EventsFilter{
		TagWhitelist: map[string][]string{"foo": {"bar"}},
	})
	if ef.matches(map[string]string{"k": "v", "foo": "bard"}) {
		t.Errorf("error matching event tags")
	}
	if !ef.matches(map[string]string{"foo": "bar"}) {
		t.Errorf("error matching event tags")
	}
}

func TestBlacklist(t *testing.T) {
	ef := newEventFilter(configuration.EventsFilter{
		TagBlacklist: map[string][]string{"foo": {"bar"}},
	})
	if !ef.matches(map[string]string{"k": "v", "foo": "bard"}) {
		t.Errorf("error matching event tags")
	}
	if ef.matches(map[string]string{"foo": "bar"}) {
		t.Errorf("error matching event tags")
	}
}

func TestWhitelistSets(t *testing.T) {
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
}

func TestBlacklistSets(t *testing.T) {
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
}
