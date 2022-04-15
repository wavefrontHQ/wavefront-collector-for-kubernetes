package control_plane

import (
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	log "github.com/sirupsen/logrus"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/configuration"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/httputil"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/summary/kubelet"
)

const (
	metricsURL    = "https://kubernetes.default.svc:443/metrics"
	metricsSource = "control_plane_source"
	jitterTime    = time.Second * 40
)

type factory struct{}

func NewFactory() factory {
	return factory{}
}

var getKubeConfigs = kubelet.GetKubeConfigs

func (p factory) Build(cfg configuration.ControlPlaneSourceConfig,
	summaryConfig configuration.SummarySourceConfig) []configuration.PrometheusSourceConfig {
	var prometheusSourceConfigs []configuration.PrometheusSourceConfig

	kubeConfig, _, err := getKubeConfigs(summaryConfig)
	if err != nil {
		log.Infof("control-plane/factory/Build: error %v", err)
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
		util.ControlplaneMetricsPrefix + "etcd.request.duration.seconds.bucket",
		util.ControlplaneMetricsPrefix + "etcd.object.counts.gauge",
		util.ControlplaneMetricsPrefix + "etcd.db.total.size.in.bytes.gauge",
		util.ControlplaneMetricsPrefix + "workqueue.adds.total.counter",
		util.ControlplaneMetricsPrefix + "workqueue.queue.duration.seconds.bucket",
	}

	promSourceConfig := p.createPrometheusSourceConfig("etcd-workqueue", httpClientConfig, metricAllowList, nil, cfg.Collection.Interval+jitterTime)
	prometheusSourceConfigs = append(prometheusSourceConfigs, promSourceConfig)

	apiServerAllowList := []string{
		util.ControlplaneMetricsPrefix + "apiserver.request.duration.seconds.bucket",
		util.ControlplaneMetricsPrefix + "apiserver.request.total.counter",
	}
	apiServerTagAllowList := map[string][]string{
		"resource": {"customresourcedefinitions", "namespaces", "lease", "nodes", "pods", "tokenreviews", "subjectaccessreviews"},
	}
	promApiServerSourceConfig := p.createPrometheusSourceConfig("apiserver", httpClientConfig, apiServerAllowList, apiServerTagAllowList, cfg.Collection.Interval)
	prometheusSourceConfigs = append(prometheusSourceConfigs, promApiServerSourceConfig)

	return prometheusSourceConfigs
}

func (p factory) createPrometheusSourceConfig(name string, httpClientConfig httputil.ClientConfig, metricAllowList []string,
	metricTagAllowList map[string][]string, collectionInterval time.Duration) configuration.PrometheusSourceConfig {

	controlPlaneTransform := configuration.Transforms{
		Source: metricsSource,
		Prefix: util.ControlplaneMetricsPrefix,
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
		Name:              name,
		UseLeaderElection: true,
	}
	return sourceConfig
}

func (p factory) BuildRuntimeConfigs(config configuration.ControlPlaneSourceConfig) []discovery.PluginConfig {
	return []discovery.PluginConfig{
		{
			Name: "coredns-discovery-controlplane",
			Type: "prometheus",
			Selectors: discovery.Selectors{
				Images: []string{"*coredns:*"},
				Labels: map[string][]string{
					"k8s-app": {"kube-dns"},
				},
			},
			Port:   "9153",
			Scheme: "http",
			Path:   "/metrics",
			Prefix: util.ControlplaneMetricsPrefix,
			Filters: filter.Config{
				MetricAllowList: []string{
					util.ControlplaneMetricsPrefix + "coredns.dns.request.duration.seconds.bucket",
					util.ControlplaneMetricsPrefix + "coredns.dns.responses.total.counter",
				},
			},
			Collection: discovery.CollectionConfig{
				Interval: config.Collection.Interval,
				Timeout:  config.Collection.Timeout,
			},
		},
	}
}
