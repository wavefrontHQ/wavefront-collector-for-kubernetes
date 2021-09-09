package cadvisor

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/prometheus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewProvider(cfg configuration.CadvisorSourceConfig, client *kubernetes.Clientset, restConfig *rest.Config) (metrics.MetricsSourceProvider, error) {
	promURLs, err := GenerateURLs(client.CoreV1().Nodes(), util.GetNodeName(), util.IsDaemonMode(), restConfig.Host)
	if err != nil {
		return nil, err
	}
	provider := &cadvisorSourceProvider{}
	for _, promURL := range promURLs {
		promSource, err := generatePrometheusSource(cfg, promURL, restConfig)
		if err != nil {
			return nil, err
		}
		provider.sources = append(provider.sources, promSource)
	}
	return provider, nil
}

type cadvisorSourceProvider struct {
	metrics.DefaultMetricsSourceProvider
	cfg     configuration.PrometheusSourceConfig
	sources []metrics.MetricsSource
}

func (c *cadvisorSourceProvider) GetMetricsSources() []metrics.MetricsSource {
	return c.sources
}

func (c *cadvisorSourceProvider) Name() string {
	return "cadvisor_metrics_provider"
}

func generatePrometheusSource(cfg configuration.CadvisorSourceConfig, promURL string, restConfig *rest.Config) (metrics.MetricsSource, error) {
	filters := filter.FromConfig(cfg.Filters)
	return prometheus.NewPrometheusMetricsSource(promURL, cfg.Prefix, cfg.Source, "", cfg.Tags, filters, generateHttpCfg(restConfig))
}

func generateHttpCfg(restConfig *rest.Config) httputil.ClientConfig {
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
