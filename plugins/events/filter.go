// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package events

import (
	"github.com/gobwas/glob"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
)

type eventFilter struct {
	whitelist     map[string]glob.Glob
	blacklist     map[string]glob.Glob
	whitelistSets []map[string]glob.Glob
	blacklistSets []map[string]glob.Glob
}

func newEventFilter(filters configuration.EventsFilter) eventFilter {
	return eventFilter{
		whitelist:     filter.MultiCompile(filters.TagWhitelist),
		blacklist:     filter.MultiCompile(filters.TagBlacklist),
		whitelistSets: filter.MultiSetCompile(filters.TagWhitelistSets),
		blacklistSets: filter.MultiSetCompile(filters.TagBlacklistSets),
	}
}

func (ef eventFilter) matches(tags map[string]string) bool {
	if len(ef.whitelist) > 0 && !filter.MatchesTags(ef.whitelist, tags) {
		return false
	}
	if len(ef.blacklist) > 0 && filter.MatchesTags(ef.blacklist, tags) {
		return false
	}

	if len(ef.whitelistSets) > 0 {
		// AND tags within a set, OR between sets
		for _, wl := range ef.whitelistSets {
			if filter.MatchesAllTags(wl, tags) {
				return true
			}
		}
		return false
	}
	if len(ef.blacklistSets) > 0 {
		// AND tags within a set, OR between sets
		for _, bl := range ef.blacklistSets {
			if filter.MatchesAllTags(bl, tags) {
				return false
			}
		}
	}
	return true
}
