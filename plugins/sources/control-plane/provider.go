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

func (c controlPlaneSourceProvider) GetMetricsSources() []metrics.Source {
	var sources []metrics.Source

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
		MetricAllowList: []string{
			"kubernertes.controlplane.etcd.request.duration.seconds.bucket",
			"kubernertes.controlplane.etcd.object.counts.gauge",
			"kubernertes.controlplane.etcd.db.total.size.in.bytes.gauge",
			"kubernertes.controlplane.workqueue.adds.total.counter",
			"kubernertes.controlplane.workqueue.queue.duration.seconds.bucket",
		},
		MetricDenyList:     nil,
		MetricTagAllowList: nil,
		MetricTagDenyList:  nil,
		TagInclude:         nil,
		TagExclude:         nil,
	})
	metricsSource, err := prometheus.NewPrometheusMetricsSource(
		"https://kubernetes.default.svc:443/metrics",
		"kubernertes.controlplane.",
		"control_plane_source",
		"",
		nil,
		controlPlaneFilters,
		httpCfg)
	if err == nil {
		sources = append(sources, metricsSource)
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
