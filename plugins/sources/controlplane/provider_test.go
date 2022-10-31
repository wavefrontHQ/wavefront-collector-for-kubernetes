package controlplane

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/options"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"k8s.io/client-go/kubernetes/fake"
)

func TestProvider(t *testing.T) {
	leadership.SetLeading(true)
	util.SetAgentType(options.AllAgentType)
	var endpoints = &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Name: "kubernetes", Namespace: "default"},
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}, {IP: "127.0.0.2"}},
			Ports:     []corev1.EndpointPort{{Name: "https", Port: 6443}},
		}},
	}

	t.Run("is identified as the correct provider", func(t *testing.T) {
		provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{}, fake.NewSimpleClientset().CoreV1())

		assert.Equal(t, "control_plane_source", provider.Name())
	})

	t.Run("builds sources for each kubernetes api server instances", func(t *testing.T) {
		provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{URL: "https://kube", InClusterConfig: "false"}, fake.NewSimpleClientset(endpoints).CoreV1())

		sources := provider.GetMetricsSources()
		assert.Equal(t, 4, len(sources), "2 prom providers querying the api x 2 instances of the api")
		assert.Equal(t, "prometheus_source: https://127.0.0.1:6443/metrics", sources[0].Name())
		assert.Equal(t, "prometheus_source: https://127.0.0.2:6443/metrics", sources[1].Name())
		assert.Equal(t, "prometheus_source: https://127.0.0.1:6443/metrics", sources[2].Name())
		assert.Equal(t, "prometheus_source: https://127.0.0.2:6443/metrics", sources[3].Name())
	})

	t.Run("implements discovery.PluginProvider", func(t *testing.T) {
		provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{URL: "https://kube", InClusterConfig: "false"}, fake.NewSimpleClientset().CoreV1())

		assert.Implements(t, (*discovery.PluginProvider)(nil), provider)
	})

	t.Run("provides one discovery plugin config for core dns", func(t *testing.T) {
		provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{URL: "https://kube", InClusterConfig: "false"}, fake.NewSimpleClientset().CoreV1())
		pluginConfigProvider := provider.(discovery.PluginProvider)

		if assert.Equal(t, 1, len(pluginConfigProvider.DiscoveryPluginConfigs())) {
			pluginConfig := pluginConfigProvider.DiscoveryPluginConfigs()[0]
			assert.Equal(t, "coredns-discovery-controlplane", pluginConfig.Name)
			assert.Equal(t, "prometheus", pluginConfig.Type)
			assert.Equal(t, metricsPrefix, pluginConfig.Prefix)
			assert.True(t, pluginConfig.Internal, "should be marked internal")
		}

	})
}
