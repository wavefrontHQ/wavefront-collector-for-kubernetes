package cadvisor

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/prometheus"
	"k8s.io/client-go/kubernetes"
	"time"
)

func NewProvider(cfg configuration.CadvisorSourceConfig, client *kubernetes.Clientset) (metrics.MetricsSourceProvider, error) {
	promURLs, err := GenerateURLs(client.CoreV1().Nodes(), util.GetNodeName(), util.IsDaemonMode())
	if err != nil {
		return nil, err
	}
	provider := &cadvisorSourceProvider{}
	for _, promURL := range promURLs {
		promSource, err := generatePrometheusSource(cfg, promURL)
		if err != nil {
			return nil, err
		}
		provider.sources = append(provider.sources, promSource)
	}
	return provider, nil
}

type cadvisorSourceProvider struct {
	cfg     configuration.PrometheusSourceConfig
	sources []metrics.MetricsSource
}

func (c *cadvisorSourceProvider) GetMetricsSources() []metrics.MetricsSource {
	return c.sources
}

func (c *cadvisorSourceProvider) Name() string {
	return "cadvisor_metrics_provider"
}

func (c *cadvisorSourceProvider) CollectionInterval() time.Duration {
	return c.cfg.Collection.Interval
}

func (c *cadvisorSourceProvider) Timeout() time.Duration {
	return c.cfg.Collection.Timeout
}

func generatePrometheusSource(cfg configuration.CadvisorSourceConfig, promURL string) (metrics.MetricsSource, error) {
	filters := filter.FromConfig(cfg.Filters)
	return prometheus.NewPrometheusMetricsSource(promURL, cfg.Prefix, cfg.Source, "", cfg.Tags, filters, cfg.HTTPClientConfig)
}
