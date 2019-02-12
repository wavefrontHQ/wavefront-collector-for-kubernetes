package discovery

import (
	"testing"
)

var sampleFile = `
global:
  discovery_interval: 5m
prom_configs:
  - name: kube-dns-discovery
    labels:
      k8s-app: kube-dns
    port: 10054
    path: /metrics
    scheme: http
    prefix: kube.dns.
    tags:
      env: prod
    filters:
      metricWhitelist:
      - '*foo*'
      - 'bar*'
      metricBlacklist:
      - 'kube.dns.go.*'
      - 'kube.dns.probe.*'
      metricTagWhitelist:
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
	if len(cfg.PromConfigs) != 1 {
		t.Errorf("error parsing yaml")
	}
	if len(cfg.PromConfigs[0].Filters.MetricWhitelist) != 2 {
		t.Errorf("error parsing filters")
	}
	if len(cfg.PromConfigs[0].Filters.MetricBlacklist) != 2 {
		t.Errorf("error parsing filters")
	}
	if len(cfg.PromConfigs[0].Filters.MetricTagWhitelist) != 2 {
		t.Errorf("error parsing filters")
	}
}
