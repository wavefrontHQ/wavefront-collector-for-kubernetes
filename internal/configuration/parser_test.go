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

prometheus_sources:
- url: 'https://kubernetes.default.svc.cluster.local:443'
  httpConfig:
    bearer_token_file: '/var/run/secrets/kubernetes.io/serviceaccount/token'
    tls_config:
      ca_file: '/var/run/secrets/kubernetes.io/serviceaccount/ca.crt'
      insecure_skip_verify: true
  prefix: 'kube.apiserver.'

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
