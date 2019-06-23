package telegraf

import (
	// Init telegraf plugins
	_ "github.com/influxdata/telegraf/plugins/inputs/cpu"
	_ "github.com/influxdata/telegraf/plugins/inputs/disk"
	_ "github.com/influxdata/telegraf/plugins/inputs/diskio"
	_ "github.com/influxdata/telegraf/plugins/inputs/kernel"
	_ "github.com/influxdata/telegraf/plugins/inputs/mem"
	_ "github.com/influxdata/telegraf/plugins/inputs/net"
	_ "github.com/influxdata/telegraf/plugins/inputs/processes"
	_ "github.com/influxdata/telegraf/plugins/inputs/swap"

	// service related plugins
	_ "github.com/influxdata/telegraf/plugins/inputs/apache"
	_ "github.com/influxdata/telegraf/plugins/inputs/consul"
	_ "github.com/influxdata/telegraf/plugins/inputs/elasticsearch"
	_ "github.com/influxdata/telegraf/plugins/inputs/memcached"
	_ "github.com/influxdata/telegraf/plugins/inputs/redis"
)
