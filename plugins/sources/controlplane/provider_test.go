package controlplane

import (
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
    "github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
    "github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
    "github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
    "testing"
)

//func mockGetKubeConfigs(cfg configuration.SummarySourceConfig) (*kube_client.Config, *kubelet.KubeletClientConfig, error) {
//	return &kube_client.Config{
//		BearerTokenFile: "tokenFile",
//		BearerToken:     "testToken",
//		TLSClientConfig: kube_client.TLSClientConfig{
//			Insecure:   false,
//			ServerName: "",
//			CertFile:   "",
//			KeyFile:    "",
//			CAFile:     "",
//			CertData:   nil,
//			KeyData:    nil,
//			CAData:     nil,
//			NextProtos: nil,
//		},
//	}, nil, nil
//}
//
//func Test_factory_Build(t *testing.T) {
//
//	getKubeConfigs = mockGetKubeConfigs
//	p := provider{}
//
//	t.Run("default", func(t *testing.T) {
//		promConfigs := p.buildPromConfigs(configuration.ControlPlaneSourceConfig{Collection: configuration.CollectionConfig{
//			Interval: time.Duration(10),
//		}}, configuration.SummarySourceConfig{
//			Transforms:        configuration.Transforms{},
//			Collection:        configuration.CollectionConfig{},
//			URL:               "https://test.com",
//			UseServiceAccount: "false",
//			Insecure:          "true",
//		})
//		assert.Equal(t, 2, len(promConfigs))
//		assert.Equal(t, "https://kubernetes.default.svc:443/metrics", promConfigs[0].URL)
//		assert.Equal(t, "testToken", promConfigs[0].HTTPClientConfig.BearerToken)
//		assert.Equal(t, time.Duration(10)+jitterTime, promConfigs[0].Collection.Interval)
//		assert.NotNil(t, promConfigs[0].Filters.MetricAllowList)
//		assert.NotNil(t, promConfigs[1].Filters.MetricAllowList)
//		assert.NotNil(t, promConfigs[1].Filters.MetricTagAllowList)
//	})
//}

func TestProvider(t *testing.T) {
	t.Run("is identified as the correct provider", func(t *testing.T) {
		provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{})

        assert.Equal(t, "control_plane_source", provider.Name())
	})

    t.Run("has two prometheus sources", func(t *testing.T) {
        provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{URL: "https://kube", InClusterConfig: "false"})

        assert.Equal(t, 2, len(provider.GetMetricsSources()))
    })

    t.Run("implements discovery.PluginConfigProvider", func(t *testing.T) {
        provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{URL: "https://kube", InClusterConfig: "false"})

        assert.Implements(t, (*discovery.PluginConfigProvider)(nil), provider)
    })

    t.Run("provides one discovery plugin config for core dns", func(t *testing.T) {
        provider, _ := NewProvider(configuration.ControlPlaneSourceConfig{}, configuration.SummarySourceConfig{URL: "https://kube", InClusterConfig: "false"})
        pluginConfigProvider := provider.(discovery.PluginConfigProvider)

        if assert.Equal(t, 1, len(pluginConfigProvider.PluginConfigs())) {
            pluginConfig := pluginConfigProvider.PluginConfigs()[0]
            assert.Equal(t, "coredns-discovery-controlplane", pluginConfig.Name)
            assert.Equal(t, "prometheus", pluginConfig.Type)
            assert.Equal(t, util.ControlplaneMetricsPrefix, pluginConfig.Prefix)
        }

    })
}

func sourceNames(sources []metrics.Source) []string {
    var names []string
    for _, source := range sources {
        names = append(names, source.Name())
    }
    return names
}
