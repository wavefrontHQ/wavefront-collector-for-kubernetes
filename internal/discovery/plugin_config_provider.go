package discovery

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
)

type PluginProvider interface {
	DiscoveryPluginConfigs(nodes util.ScrapeNodes) []PluginConfig
}
