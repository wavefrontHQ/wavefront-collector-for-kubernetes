package control_plane

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/summary/kubelet"
	kube_client "k8s.io/client-go/rest"
)

func mockGetKubeConfigs(cfg configuration.SummarySourceConfig) (*kube_client.Config, *kubelet.KubeletClientConfig, error) {
	return &kube_client.Config{
		BearerTokenFile: "tokenFile",
		BearerToken:     "testToken",
		TLSClientConfig: kube_client.TLSClientConfig{
			Insecure:   false,
			ServerName: "",
			CertFile:   "",
			KeyFile:    "",
			CAFile:     "",
			CertData:   nil,
			KeyData:    nil,
			CAData:     nil,
			NextProtos: nil,
		},
	}, nil, nil
}

func Test_factory_Build(t *testing.T) {

	getKubeConfigs = mockGetKubeConfigs
	p := factory{}

	t.Run("default", func(t *testing.T) {
		promConfigs := p.Build(configuration.ControlPlaneSourceConfig{Collection: configuration.CollectionConfig{
			Interval: time.Duration(10),
		}}, configuration.SummarySourceConfig{
			Transforms:        configuration.Transforms{},
			Collection:        configuration.CollectionConfig{},
			URL:               "https://test.com",
			UseServiceAccount: "false",
			Insecure:          "true",
		})
		assert.Equal(t, 2, len(promConfigs))
		assert.Equal(t, "https://kubernetes.default.svc:443/metrics", promConfigs[0].URL)
		assert.Equal(t, "testToken", promConfigs[0].HTTPClientConfig.BearerToken)
		assert.Equal(t, time.Duration(10)+jitterTime, promConfigs[0].Collection.Interval)
		assert.NotNil(t, promConfigs[0].Filters.MetricAllowList)
		assert.NotNil(t, promConfigs[1].Filters.MetricAllowList)
		assert.NotNil(t, promConfigs[1].Filters.MetricTagAllowList)
	})
}
