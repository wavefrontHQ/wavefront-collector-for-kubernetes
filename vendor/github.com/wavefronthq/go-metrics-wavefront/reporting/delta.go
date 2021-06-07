package reporting

import (
	"strings"
	"unicode/utf8"
)

var (
	deltaPrefix           = "\u2206"
	altDeltaPrefix        = "\u0394"
	_, deltaPrefixSize    = utf8.DecodeRuneInString(deltaPrefix)
	_, altDeltaPrefixSize = utf8.DecodeRuneInString(altDeltaPrefix)
)

// DeltaCounterName return a delta counter name prefixed with ∆.
// Can be used as an input for RegisterMetric() or GetOrRegisterMetric() functions
func DeltaCounterName(name string) string {
	if hasDeltaPrefix(name) {
		return name
	}
	return deltaPrefix + name
}

func hasDeltaPrefix(name string) bool {
	return strings.HasPrefix(name, deltaPrefix) || strings.HasPrefix(name, altDeltaPrefix)
}
