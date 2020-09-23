// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package events

import (
	"github.com/gobwas/glob"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
)

type eventFilter struct {
	allowList     map[string]glob.Glob
	denyList      map[string]glob.Glob
	allowListSets []map[string]glob.Glob
	denyListSets  []map[string]glob.Glob
}

func newEventFilter(filters configuration.EventsFilter) eventFilter {
	allowList := filters.TagWhitelist
	if len(filters.TagAllowList) > 0 {
		allowList = filters.TagAllowList
	}
	denyList := filters.TagBlacklist
	if len(filters.TagDenyList) > 0 {
		denyList = filters.TagDenyList
	}
	allowListSets := filters.TagWhitelistSets
	if len(filters.TagAllowListSets) > 0 {
		allowListSets = filters.TagAllowListSets
	}
	denyListSets := filters.TagBlacklistSets
	if len(filters.TagDenyListSets) > 0 {
		denyListSets = filters.TagDenyListSets
	}

	return eventFilter{
		allowList:     filter.MultiCompile(allowList),
		denyList:      filter.MultiCompile(denyList),
		allowListSets: filter.MultiSetCompile(allowListSets),
		denyListSets:  filter.MultiSetCompile(denyListSets),
	}
}

func (ef eventFilter) matches(tags map[string]string) bool {
	if len(ef.allowList) > 0 && !filter.MatchesTags(ef.allowList, tags) {
		return false
	}
	if len(ef.denyList) > 0 && filter.MatchesTags(ef.denyList, tags) {
		return false
	}

	if len(ef.allowListSets) > 0 {
		// AND tags within a set, OR between sets
		for _, wl := range ef.allowListSets {
			if filter.MatchesAllTags(wl, tags) {
				return true
			}
		}
		return false
	}
	if len(ef.denyListSets) > 0 {
		// AND tags within a set, OR between sets
		for _, bl := range ef.denyListSets {
			if filter.MatchesAllTags(bl, tags) {
				return false
			}
		}
	}
	return true
}
