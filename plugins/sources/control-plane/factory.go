package control_plane

import (
    log "github.com/sirupsen/logrus"
    "time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/summary/kubelet"
)

const (
	metricsURL    = "https://kubernetes.default.svc:443/metrics"
	metricsPrefix = "kubernetes.controlplane."
	metricsSource = "control_plane_source"
)

type factory struct{}

// Returns a new prometheus provider factory
func NewFactory() factory {
	return factory{}
}

func (p factory) Build(cfg configuration.ControlPlaneSourceConfig,
	summaryConfig configuration.SummarySourceConfig) []configuration.PrometheusSourceConfig {
	var prometheusSourceConfigs []configuration.PrometheusSourceConfig

	kubeConfig, _, err := kubelet.GetKubeConfigs(summaryConfig)
	if err != nil {
	    log.Infof("control-plane/factory/Build: error %s", err.Error())
		return nil
	}
	httpClientConfig := httputil.ClientConfig{
		BearerTokenFile: kubeConfig.BearerTokenFile,
		BearerToken:     kubeConfig.BearerToken,
		TLSConfig: httputil.TLSConfig{
			CAFile:             kubeConfig.CAFile,
			CertFile:           kubeConfig.CertFile,
			KeyFile:            kubeConfig.KeyFile,
			ServerName:         kubeConfig.ServerName,
			InsecureSkipVerify: kubeConfig.Insecure,
		},
	}
	metricAllowList := []string{
		"kubernetes.controlplane.etcd.request.duration.seconds.bucket",
		"kubernetes.controlplane.etcd.object.counts.gauge",
		"kubernetes.controlplane.etcd.db.total.size.in.bytes.gauge",
		"kubernetes.controlplane.workqueue.adds.total.counter",
		"kubernetes.controlplane.workqueue.queue.duration.seconds.bucket",
	}

	promSourceConfig := p.createPrometheusSourceConfig(httpClientConfig, metricAllowList, nil, cfg.Collection.Interval+40*time.Second)
	prometheusSourceConfigs = append(prometheusSourceConfigs, promSourceConfig)

	apiServerAllowList := []string{
		"kubernetes.controlplane.apiserver.request.duration.seconds.bucket",
		"kubernetes.controlplane.apiserver.request.total.counter",
	}
	apiServerTagAllowList := map[string][]string{
		"resource": {"customresourcedefinitions", "namespaces", "lease", "nodes", "pods", "tokenreviews", "subjectaccessreviews"},
	}
	promApiServerSourceConfig := p.createPrometheusSourceConfig(httpClientConfig, apiServerAllowList, apiServerTagAllowList, cfg.Collection.Interval)
	prometheusSourceConfigs = append(prometheusSourceConfigs, promApiServerSourceConfig)

	return prometheusSourceConfigs
}

func (p factory) createPrometheusSourceConfig(httpClientConfig httputil.ClientConfig, metricAllowList []string,
	metricTagAllowList map[string][]string, collectionInterval time.Duration) configuration.PrometheusSourceConfig {

	controlPlaneTransform := configuration.Transforms{
		Source: metricsSource,
		Prefix: metricsPrefix,
		Tags:   nil,
		Filters: filter.Config{
			MetricAllowList:    metricAllowList,
			MetricDenyList:     nil,
			MetricTagAllowList: metricTagAllowList,
			MetricTagDenyList:  nil,
			TagInclude:         nil,
			TagExclude:         nil,
		},
	}

	sourceConfig := configuration.PrometheusSourceConfig{
		Transforms: controlPlaneTransform,
		Collection: configuration.CollectionConfig{
			Interval: collectionInterval,
			Timeout:  0,
		},
		URL:               metricsURL,
		HTTPClientConfig:  httpClientConfig,
		Discovered:        "",
		Name:              "",
		UseLeaderElection: false,
	}
	return sourceConfig
}
