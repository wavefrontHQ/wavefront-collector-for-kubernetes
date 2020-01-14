// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package systemd

import (
	"github.com/gobwas/glob"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
)

type unitFilter struct {
	unitWhitelist glob.Glob
	unitBlacklist glob.Glob
}

func (uf *unitFilter) match(name string) bool {
	if uf.unitWhitelist != nil && !uf.unitWhitelist.Match(name) {
		return false
	}
	if uf.unitBlacklist != nil && uf.unitBlacklist.Match(name) {
		return false
	}
	return true
}

func fromConfig(whitelist, blacklist []string) *unitFilter {
	if len(whitelist) == 0 && len(blacklist) == 0 {
		return nil
	}
	return &unitFilter{
		unitWhitelist: filter.Compile(whitelist),
		unitBlacklist: filter.Compile(blacklist),
	}
}
