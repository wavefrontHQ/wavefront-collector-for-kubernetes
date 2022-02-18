// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"testing"
)

var sampleFile = `
global:
  discovery_interval: 5m
plugins:
  - type: telegraf/redis
    name: "redis"
    selectors:
      images:
      - 'redis:*'
      - '*redis*'
    port: 6379
    scheme: "tcp"
    conf: |
      servers = [${server}]
      password = bar
  - type: telegraf/memcached
    name: "memcached"
    selectors:
      images:
      - 'memcached:*'
    port: 11211
    conf: |
      servers = ${server}
  - type: prometheus
    name: kube-dns
    selectors:
      labels:
        k8s-app: 
        - kube-dns      
    port: 10054
    path: /metrics
    scheme: http
    prefix: kubernetes.dns.
    tags:
      env: prod
    filters:
      metricAllowList:
      - '*foo*'
      - 'bar*'
      metricDenyList:
      - 'kubernetes.dns.go.*'
      - 'kubernetes.dns.probe.*'
      metricTagAllowList:
        env:
        - 'prod1*'
        - 'prod2*'
        service:
        - 'app1*'
        - '?app2*'
`

func TestFromYAML(t *testing.T) {
	cfg, err := FromYAML([]byte(sampleFile))
	if err != nil {
		t.Errorf("error loading yaml: %q", err)
		return
	}
	if len(cfg.PluginConfigs) != 3 {
		t.Errorf("invalid number of plugins")
	}
	if len(cfg.PluginConfigs[2].Filters.MetricAllowList) != 2 {
		t.Errorf("error parsing filters")
	}
	if len(cfg.PluginConfigs[2].Filters.MetricDenyList) != 2 {
		t.Errorf("error parsing filters")
	}
	if len(cfg.PluginConfigs[2].Filters.MetricTagAllowList) != 2 {
		t.Errorf("error parsing filters")
	}
	if len(cfg.PluginConfigs[0].Selectors.Images) != 2 {
		t.Errorf("error parsing plugin images")
	}
}
