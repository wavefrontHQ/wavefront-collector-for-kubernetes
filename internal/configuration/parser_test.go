package configuration

import (
	"testing"
)

var sampleFile = `
clusterName: new-collector
enableDiscovery: true

sinks:
- proxyAddress: wavefront-proxy.default.svc.cluster.local:2878
  tags:
    env: gcp-dev
    image: 0.9.9-rc3

kubernetes_source:
  prefix: kubernetes.

discovery_configs:
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
`

func TestFromYAML(t *testing.T) {
	cfg, err := FromYAML([]byte(sampleFile))
	if err != nil {
		t.Errorf("error loading yaml: %q", err)
		return
	}
	opt, err := cfg.Convert()
	if err != nil {
		t.Errorf("error converting cfg: %q", err)
	}

	if len(opt.Sources) == 0 {
		t.Errorf("error in converting")
	}
}
