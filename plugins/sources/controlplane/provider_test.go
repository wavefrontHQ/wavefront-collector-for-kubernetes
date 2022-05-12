package controlplane

import (
	"testing"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/leadership"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
)

func TestProvider(t *testing.T) {
	leadership.SetLeading(true)
	_ = util.SetScrapeCluster(true)

	t.Run("is identified as the correct provider", func(t *testing.T) {
		provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{})

		assert.Equal(t, "control_plane_source", provider.Name())
	})

	t.Run("has two prometheus sources", func(t *testing.T) {
		provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{URL: "https://kube", InClusterConfig: "false"})

		assert.Equal(t, 2, len(provider.GetMetricsSources()))
	})

	t.Run("implements discovery.PluginProvider", func(t *testing.T) {
		provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{URL: "https://kube", InClusterConfig: "false"})

		assert.Implements(t, (*discovery.PluginProvider)(nil), provider)
	})

	t.Run("provides one discovery plugin config for core dns", func(t *testing.T) {
		provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{URL: "https://kube", InClusterConfig: "false"})
		pluginConfigProvider := provider.(discovery.PluginProvider)

		if assert.Equal(t, 1, len(pluginConfigProvider.DiscoveryPluginConfigs(util.ScrapeNodes{Value: "own"}))) {
			pluginConfig := pluginConfigProvider.DiscoveryPluginConfigs(util.ScrapeNodes{Value: "own"})[0]
			assert.Equal(t, "coredns-discovery-controlplane", pluginConfig.Name)
			assert.Equal(t, "prometheus", pluginConfig.Type)
			assert.Equal(t, metricsPrefix, pluginConfig.Prefix)
			assert.True(t, pluginConfig.Internal, "should be marked internal")
		}

	})
}
