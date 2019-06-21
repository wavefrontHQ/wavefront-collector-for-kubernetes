package redis

import (
	"fmt"

	wfTelegraf "github.com/wavefronthq/wavefront-kubernetes-collector/internal/telegraf"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/redis"
)

type handler struct{}

func NewPluginHandler() wfTelegraf.PluginHandler {
	return handler{}
}

func (r handler) Init(input telegraf.Input, vals map[string][]string) error {
	redisPlugin, ok := input.(*redis.Redis)
	if !ok {
		return fmt.Errorf("invalid input type: %s", input.Description())
	}

	if len(vals["servers"]) > 0 {
		redisPlugin.Servers = []string{vals["servers"][0]}
	} else {
		return fmt.Errorf("missing redis server address")
	}
	//TODO: support password and tlsConfig
	return nil
}
