package control_plane

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/prometheus"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/summary/kubelet"
	"k8s.io/client-go/rest"
)

type controlPlaneSourceProvider struct {
	metrics.DefaultSourceProvider
	config     configuration.ControlPlaneSourceConfig
	kubeConfig *rest.Config
}

const (
	metricsURL    = "https://kubernetes.default.svc:443/metrics"
	metricsPrefix = "kubernetes.controlplane."
	metricsSource = "control_plane_source"
)

func (c controlPlaneSourceProvider) newPrometheusMetricSource(metricAllowList []string, metricTagAllowList map[string][]string) (metrics.Source, error) {
	httpCfg := httputil.ClientConfig{
		BearerTokenFile: c.kubeConfig.BearerTokenFile,
		BearerToken:     c.kubeConfig.BearerToken,
		TLSConfig: httputil.TLSConfig{
			CAFile:             c.kubeConfig.CAFile,
			CertFile:           c.kubeConfig.CertFile,
			KeyFile:            c.kubeConfig.KeyFile,
			ServerName:         c.kubeConfig.ServerName,
			InsecureSkipVerify: c.kubeConfig.Insecure,
		},
	}

	controlPlaneFilters := filter.NewGlobFilter(filter.Config{
		MetricAllowList:    metricAllowList,
		MetricDenyList:     nil,
		MetricTagAllowList: metricTagAllowList,
		MetricTagDenyList:  nil,
		TagInclude:         nil,
		TagExclude:         nil,
	})
	metricsSource, err := prometheus.NewPrometheusMetricsSource(
		metricsURL,
		metricsPrefix,
		metricsSource,
		"",
		nil,
		controlPlaneFilters,
		httpCfg)

	return metricsSource, err
}

func (c controlPlaneSourceProvider) GetMetricsSources() []metrics.Source {
	var sources []metrics.Source

	metricAllowList := []string{
		"kubernetes.controlplane.etcd.request.duration.seconds.bucket",
		"kubernetes.controlplane.etcd.object.counts.gauge",
		"kubernetes.controlplane.etcd.db.total.size.in.bytes.gauge",
		"kubernetes.controlplane.workqueue.adds.total.counter",
		"kubernetes.controlplane.workqueue.queue.duration.seconds.bucket",
	}
	metricsSource, err := c.newPrometheusMetricSource(metricAllowList, nil)
	if err == nil {
		sources = append(sources, metricsSource)
	} else {
		return nil
	}

	apiServerAllowList := []string{
		"kubernetes.controlplane.apiserver.request.duration.seconds.bucket",
		"kubernetes.controlplane.apiserver.request.total.counter",
	}
	apiServerTagAllowList := map[string][]string{
		"resource": {"customresourcedefinitions", "namespaces", "lease", "nodes", "pods", "tokenreviews", "subjectaccessreviews"},
	}
	apiServerSource, err := c.newPrometheusMetricSource(apiServerAllowList, apiServerTagAllowList)
	if err == nil {
		sources = append(sources, apiServerSource)
	} else {
		return nil
	}

	return sources
}

func (c controlPlaneSourceProvider) Name() string {
	return "control_plane_metrics_provider"
}

func NewProvider(
	config configuration.ControlPlaneSourceConfig,
	summaryConfig configuration.SummarySourceConfig,
) (metrics.SourceProvider, error) {
	kubeConfig, _, err := kubelet.GetKubeConfigs(summaryConfig)
	if err != nil {
		return nil, err
	}
	return &controlPlaneSourceProvider{
		config:     config,
		kubeConfig: kubeConfig,
	}, nil
}
