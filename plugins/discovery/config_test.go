// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"

	"github.com/stretchr/testify/assert"

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
	configResource := makeConfigResource("test", map[string]string{"plugins": sampleConfigString})
	cfg, err := load(configResource.data)
	assert.NoError(t, err)
	assert.True(t, len(cfg.PluginConfigs) == 3)
}

func TestCombine(t *testing.T) {
	cfg := &discovery.Config{
		DiscoveryInterval:          time.Duration(10 * time.Minute),
		AnnotationPrefix:           "wavefront.com",
		PluginConfigs:              makePlugins(3, "main"),
		DisableAnnotationDiscovery: true,
	}

	plugins := map[string]discovery.Config{
		"memcached": {PluginConfigs: makePlugins(3, "memcached")},
		"redis":     {PluginConfigs: makePlugins(2, "redis")},
	}

	result := combine(*cfg, plugins)
	assert.Equal(t, 8, len(result.PluginConfigs))
	assert.True(t, result.DisableAnnotationDiscovery)

	result = combine(*cfg, nil)
	assert.Equal(t, 3, len(result.PluginConfigs))

	cfg.PluginConfigs = nil
	result = combine(*cfg, plugins)
	assert.Equal(t, 5, len(result.PluginConfigs))
}

func TestAdd(t *testing.T) {
	ch := makeFakeConfigHandler(discovery.Config{PluginConfigs: makePlugins(2, "main")})
	configResource := makeConfigResource("test", map[string]string{"plugins": sampleConfigString})

	// no changes as missing annotation on configResource
	ch.updated(configResource)
	assert.True(t, ch.changed == false)

	// changes when annotation is present
	configResource.meta.SetAnnotations(map[string]string{discoveryAnnotation: "true"})
	ch.updated(configResource)
	assert.True(t, ch.changed)

	cfg, changed := ch.Config()
	assert.True(t, changed)
	assert.Equal(t, 5, len(cfg.PluginConfigs))
}

func TestDelete(t *testing.T) {
	ch := makeFakeConfigHandler(discovery.Config{PluginConfigs: makePlugins(2, "main")})
	configResource := makeConfigResource("test", map[string]string{"plugins": sampleConfigString})
	configResource.meta.SetAnnotations(map[string]string{discoveryAnnotation: "true"})
	ch.updated(configResource)

	cfg, changed := ch.Config()
	assert.True(t, changed)
	assert.Equal(t, 5, len(cfg.PluginConfigs))

	// delete the config map and validate the plugins are removed
	ch.deleted(configResource.meta.Name)
	cfg, changed = ch.Config()
	assert.True(t, changed)
	assert.Equal(t, 2, len(cfg.PluginConfigs))
}

func TestConvertByteArrayData(t *testing.T) {
	secretData := map[string][]byte{"plugins": []byte(sampleConfigString)}
	convertedData := convertByteArrayData(secretData)
	assert.Equal(t, map[string]string{"plugins": sampleConfigString}, convertedData)
}

func TestUpdateConfigIfValidDoesntPanic(t *testing.T) {
	fakeConfigHandler := makeFakeConfigHandler(discovery.Config{PluginConfigs: makePlugins(2, "main")})
	configResource := makeConfigMap("test", map[string]string{"plugins": sampleConfigString})

	obj := cache.DeletedFinalStateUnknown{Key: "bar", Obj: configResource}

	assert.NotPanics(t, func() { updateConfigMapIfValid(obj, fakeConfigHandler) }, "updateConfigMapIfValid panicked")
}

func TestDeleteConfigIfValidDoesntPanic(t *testing.T) {
	fakeConfigHandler := makeFakeConfigHandler(discovery.Config{PluginConfigs: makePlugins(2, "main")})
	configResource := makeConfigMap("test", map[string]string{"plugins": sampleConfigString})

	obj := cache.DeletedFinalStateUnknown{Key: "bar", Obj: configResource}

	assert.NotPanics(t, func() { deleteConfigMapIfValid(obj, fakeConfigHandler) }, "deleteConfigMapIfValid panicked")
}

func TestUpdateSecretIfValidDoesntPanic(t *testing.T) {
	fakeConfigHandler := makeFakeConfigHandler(discovery.Config{PluginConfigs: makePlugins(2, "main")})
	configResource := makeSecret("test", map[string]string{"plugins": sampleConfigString})

	obj := cache.DeletedFinalStateUnknown{Key: "bar", Obj: configResource}

	assert.NotPanics(t, func() { updateSecretIfValid(obj, fakeConfigHandler) }, "updateConfigMapIfValid panicked")
}

func TestDeleteSecretIfValidDoesntPanic(t *testing.T) {
	fakeConfigHandler := makeFakeConfigHandler(discovery.Config{PluginConfigs: makePlugins(2, "main")})
	configResource := makeSecret("test", map[string]string{"plugins": sampleConfigString})

	obj := cache.DeletedFinalStateUnknown{Key: "bar", Obj: configResource}

	assert.NotPanics(t, func() { deleteSecretIfValid(obj, fakeConfigHandler) }, "deleteConfigMapIfValid panicked")
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
			Name: prefix + "-plugin-" + fmt.Sprint(i),
			Type: "prometheus",
		}
	}
	return plugins
}

func makeConfigResource(name string, data map[string]string) *configResource {
	return &configResource{
		meta: metav1.ObjectMeta{
			Name: name,
		},
		data: data,
	}
}

func makeConfigMap(name string, data map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: data,
	}
}

func makeSecret(name string, data map[string]string) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		StringData: data,
	}
}
