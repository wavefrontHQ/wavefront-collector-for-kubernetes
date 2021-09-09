// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var sampleFile = `
clusterName: new-collector
enableDiscovery: true
enableEvents: true
defaultCollectionInterval: 10s
omitBucketSuffix: true

sinks:
- proxyAddress: wavefront-proxy.default.svc.cluster.local:2878
  tags:
    env: gcp-dev
    image: 0.9.9-rc3
  filters:
    metricAllowList:
    - 'kubernetes.node.*'

    metricTagAllowList:
      nodename:
      - 'gke-vikramr-cluster*wj2d'

    tagInclude:
    - 'nodename'

events:
  filters:
    tagAllowList:
      namespace:
      - "default"
      component:
      - "pp"
    tagAllowListSets:
    - kind:
      - "Pod"
      reason:
      - "Scheduled"
    - kind:
      - "DaemonSet"
      reason:
      - "SuccessfulCreate"

sources:
  kubernetes_source:
    prefix: kubernetes.

  kubernetes_cadvisor_source:
    prefix: 'kubernetes.cadvisor.'
    perNode: true
    filters:
      metricAllowList:
      - 'kubernetes.cadvisor.*'

  prometheus_sources:
  - url: 'https://kubernetes.default.svc.cluster.local:443'
    httpConfig:
      bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token'
      tls_config:
        ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt'
        insecure_skip_verify: true
    prefix: 'kube.apiserver.'

  telegraf_sources:
    - plugins: [cpu]
      collection:
        interval: 1s
    - plugins: [mem]

discovery:
  plugins:
  - type: telegraf/redis
    name: "redis"
    selectors:
      images:
      - 'redis:*'
      - '*redis*'
    port: 6379
    scheme: "tcp"
    collection:
      interval: 1s
    conf: |
      servers = [${server}]
      password = bar
`

func TestFromYAML(t *testing.T) {
	cfg, err := FromYAML([]byte(sampleFile))
	if err != nil {
		t.Errorf("error loading yaml: %q", err)
		return
	}
	if len(cfg.Sinks) == 0 {
		t.Errorf("invalid sinks")
	}

	assert.True(t, cfg.EnableEvents)
	assert.Equal(t, "default", cfg.EventsConfig.Filters.TagAllowList["namespace"][0])
	assert.Equal(t, "pp", cfg.EventsConfig.Filters.TagAllowList["component"][0])

	assert.True(t, len(cfg.Sources.PrometheusConfigs) > 0)
	assert.Equal(t, "kubernetes.", cfg.Sources.SummaryConfig.Prefix)
	assert.Equal(t, "kube.apiserver.", cfg.Sources.PrometheusConfigs[0].Prefix)
	assert.Equal(t, "kubernetes.cadvisor.", cfg.Sources.CadvisorConfig.Prefix)
}
