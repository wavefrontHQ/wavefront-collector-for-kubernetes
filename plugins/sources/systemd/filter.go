package systemd

import (
	"github.com/gobwas/glob"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
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

func fromQuery(vals map[string][]string) *unitFilter {
	if len(vals) == 0 {
		return nil
	}
	unitWhitelist := vals["unitWhitelist"]
	unitBlacklist := vals["unitBlacklist"]

	return &unitFilter{
		unitWhitelist: filter.Compile(unitWhitelist),
		unitBlacklist: filter.Compile(unitBlacklist),
	}
}
