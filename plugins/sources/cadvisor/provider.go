package cadvisor

import (
	log "github.com/sirupsen/logrus"
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

type cadvisorSourceProvider struct {
	metrics.DefaultMetricsSourceProvider
	config        configuration.CadvisorSourceConfig
	k8sClient     *kubernetes.Clientset
	k8sConfig     *rest.Config
	kubeletConfig *kubelet.KubeletClientConfig
}

func NewProvider(
	cfg configuration.CadvisorSourceConfig,
	client *kubernetes.Clientset,
	restConfig *rest.Config,
	kubeletConfig *kubelet.KubeletClientConfig,
) metrics.MetricsSourceProvider {
	return &cadvisorSourceProvider{
		config:        cfg,
		k8sClient:     client,
		k8sConfig:     restConfig,
		kubeletConfig: kubeletConfig,
	}
}

func (c *cadvisorSourceProvider) GetMetricsSources() []metrics.MetricsSource {
	promURLs, err := GenerateURLs(c.k8sClient.CoreV1().Nodes(), util.GetNodeName(), util.IsDaemonMode(), c.kubeletConfig.BaseURL)
	if err != nil {
		log.Errorf("error getting sources for cAdvisor: %s", err.Error())
		return nil
	}
	var sources []metrics.MetricsSource
	for _, promURL := range promURLs {
		promSource, err := generatePrometheusSource(c.config, promURL.String(), c.k8sConfig)
		if err != nil {
			log.Errorf("error generating sources for cAdvisor: %s", err.Error())
			return nil
		}
		sources = append(sources, promSource)
	}
	return sources
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
