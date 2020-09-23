// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package systemd

import (
	"github.com/gobwas/glob"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
)

type unitFilter struct {
	unitAllowList glob.Glob
	unitDenyList  glob.Glob
}

func (uf *unitFilter) match(name string) bool {
	if uf.unitAllowList != nil && !uf.unitAllowList.Match(name) {
		return false
	}
	if uf.unitDenyList != nil && uf.unitDenyList.Match(name) {
		return false
	}
	return true
}

func fromConfig(allowList, denyList []string) *unitFilter {
	if len(allowList) == 0 && len(denyList) == 0 {
		return nil
	}
	return &unitFilter{
		unitAllowList: filter.Compile(allowList),
		unitDenyList:  filter.Compile(denyList),
	}
}
