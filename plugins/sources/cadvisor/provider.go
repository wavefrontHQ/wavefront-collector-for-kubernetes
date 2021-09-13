package cadvisor

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/prometheus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/summary/kubelet"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewProvider(cfg configuration.CadvisorSourceConfig, client *kubernetes.Clientset, restConfig *rest.Config, kubeletConfig *kubelet.KubeletClientConfig) (metrics.MetricsSourceProvider, error) {
	promURLs, err := GenerateURLs(client.CoreV1().Nodes(), util.GetNodeName(), util.IsDaemonMode(), kubeletConfig.BaseURL)
	if err != nil {
		return nil, err
	}
	provider := &cadvisorSourceProvider{}
	for _, promURL := range promURLs {
		promSource, err := generatePrometheusSource(cfg, promURL.String(), restConfig)
		if err != nil {
			return nil, err
		}
		provider.sources = append(provider.sources, promSource)
	}
	return provider, nil
}

type cadvisorSourceProvider struct {
	metrics.DefaultMetricsSourceProvider
	sources []metrics.MetricsSource
}

func (c *cadvisorSourceProvider) GetMetricsSources() []metrics.MetricsSource {
	return c.sources
}

func (c *cadvisorSourceProvider) Name() string {
	return "cadvisor_metrics_provider"
}

func generatePrometheusSource(cfg configuration.CadvisorSourceConfig, promURL string, restConfig *rest.Config) (metrics.MetricsSource, error) {
	return prometheus.NewPrometheusMetricsSource(
		promURL,
		cfg.Prefix,
		cfg.Source,
		"",
		cfg.Tags,
		filter.FromConfig(cfg.Filters),
		generateHTTPCfg(restConfig),
	)
}

func generateHTTPCfg(restConfig *rest.Config) httputil.ClientConfig {
	return httputil.ClientConfig{
		BearerTokenFile: restConfig.BearerTokenFile,
		BearerToken:     restConfig.BearerToken,
		TLSConfig: httputil.TLSConfig{
			CAFile:             restConfig.CAFile,
			CertFile:           restConfig.CertFile,
			KeyFile:            restConfig.KeyFile,
			ServerName:         restConfig.ServerName,
			InsecureSkipVerify: restConfig.Insecure,
		},
	}
}
