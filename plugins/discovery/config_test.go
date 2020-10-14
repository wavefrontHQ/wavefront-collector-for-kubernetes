package discovery

import (
	"testing"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"

	"github.com/stretchr/testify/assert"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var sampleConfigString = `
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
    prefix: kube.dns.
    tags:
      env: prod
    filters:
      metricAllowList:
      - '*foo*'
      - 'bar*'
      metricDenyList:
      - 'kube.dns.go.*'
      - 'kube.dns.probe.*'
      metricTagAllowList:
        env:
        - 'prod1*'
        - 'prod2*'
        service:
        - 'app1*'
        - '?app2*'
`

func TestLoad(t *testing.T) {
	cmap := makeCfgMap("test", map[string]string{"plugins": sampleConfigString})
	cfg, err := load(cmap)
	assert.NoError(t, err)
	assert.True(t, len(cfg.PluginConfigs) == 3)
}

func TestCombine(t *testing.T) {
	cfg := &discovery.Config{
		DiscoveryInterval: time.Duration(10 * time.Minute),
		AnnotationPrefix:  "wavefront.com",
		PluginConfigs:     makePlugins(3, "main"),
	}

	plugins := map[string]discovery.Config{
		"memcached": {PluginConfigs: makePlugins(3, "memcached")},
		"redis":     {PluginConfigs: makePlugins(2, "redis")},
	}
	result := combine(*cfg, plugins)
	assert.Equal(t, 8, len(result.PluginConfigs))

	result = combine(*cfg, nil)
	assert.Equal(t, 3, len(result.PluginConfigs))

	cfg.PluginConfigs = nil
	result = combine(*cfg, plugins)
	assert.Equal(t, 5, len(result.PluginConfigs))
}

func TestAdd(t *testing.T) {
	ch := makeFakeConfigHandler(discovery.Config{PluginConfigs: makePlugins(2, "main")})
	cmap := makeCfgMap("test", map[string]string{"plugins": sampleConfigString})

	// no changes as missing annotation on cmap
	ch.updated(cmap)
	assert.True(t, ch.changed == false)

	// changes when annotation is present
	cmap.SetAnnotations(map[string]string{discoveryAnnotation: "true"})
	ch.updated(cmap)
	assert.True(t, ch.changed)

	cfg, changed := ch.Config()
	assert.True(t, changed)
	assert.Equal(t, 5, len(cfg.PluginConfigs))
}

func TestDelete(t *testing.T) {
	ch := makeFakeConfigHandler(discovery.Config{PluginConfigs: makePlugins(2, "main")})
	cmap := makeCfgMap("test", map[string]string{"plugins": sampleConfigString})
	cmap.SetAnnotations(map[string]string{discoveryAnnotation: "true"})
	ch.updated(cmap)

	cfg, changed := ch.Config()
	assert.True(t, changed)
	assert.Equal(t, 5, len(cfg.PluginConfigs))

	// delete the config map and validate the plugins are removed
	ch.deleted(cmap)
	cfg, changed = ch.Config()
	assert.True(t, changed)
	assert.Equal(t, 2, len(cfg.PluginConfigs))
}

func makeFakeConfigHandler(wired discovery.Config) *configHandler {
	return &configHandler{
		wiredCfg:    wired,
		runtimeCfgs: make(map[string]discovery.Config),
	}
}

func makePlugins(n int, prefix string) []discovery.PluginConfig {
	plugins := make([]discovery.PluginConfig, n)
	for i := 0; i < n; i++ {
		plugins[0] = discovery.PluginConfig{
			Name: prefix + "-plugin-" + string(rune(i)),
			Type: "prometheus",
		}
	}
	return plugins
}

func makeCfgMap(name string, data map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: data,
	}
}
