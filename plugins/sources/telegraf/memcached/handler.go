package memcached

import (
	"fmt"

	wfTelegraf "github.com/wavefronthq/wavefront-kubernetes-collector/internal/telegraf"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/memcached"
)

type handler struct{}

func NewPluginHandler() wfTelegraf.PluginHandler {
	return handler{}
}

func (r handler) Init(input telegraf.Input, vals map[string][]string) error {
	memcachedPlugin, ok := input.(*memcached.Memcached)
	if !ok {
		return fmt.Errorf("invalid input type: %s", input.Description())
	}

	if len(vals["server"]) > 0 {
		memcachedPlugin.Servers = []string{vals["server"][0]}
	} else {
		return fmt.Errorf("missing memcached server address")
	}
	return nil
}
