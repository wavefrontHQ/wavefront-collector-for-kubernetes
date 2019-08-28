// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

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
	_ "github.com/influxdata/telegraf/plugins/inputs/activemq"
	_ "github.com/influxdata/telegraf/plugins/inputs/apache"
	_ "github.com/influxdata/telegraf/plugins/inputs/consul"
	_ "github.com/influxdata/telegraf/plugins/inputs/couchbase"
	_ "github.com/influxdata/telegraf/plugins/inputs/couchdb"
	_ "github.com/influxdata/telegraf/plugins/inputs/elasticsearch"
	_ "github.com/influxdata/telegraf/plugins/inputs/haproxy"
	_ "github.com/influxdata/telegraf/plugins/inputs/jolokia2"
	_ "github.com/influxdata/telegraf/plugins/inputs/memcached"
	_ "github.com/influxdata/telegraf/plugins/inputs/mongodb"
	_ "github.com/influxdata/telegraf/plugins/inputs/mysql"
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx"
	_ "github.com/influxdata/telegraf/plugins/inputs/nginx_plus"
	_ "github.com/influxdata/telegraf/plugins/inputs/postgresql"
	_ "github.com/influxdata/telegraf/plugins/inputs/rabbitmq"
	_ "github.com/influxdata/telegraf/plugins/inputs/redis"
	_ "github.com/influxdata/telegraf/plugins/inputs/riak"
	_ "github.com/influxdata/telegraf/plugins/inputs/zookeeper"
)
