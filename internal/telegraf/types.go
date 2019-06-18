package telegraf

import (
	"github.com/influxdata/telegraf"
)

type PluginHandler interface {
	Init(input telegraf.Input, vals map[string][]string) error
}
