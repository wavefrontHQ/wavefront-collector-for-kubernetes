package controlplane

import (
	"fmt"
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/util"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/discovery"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/metrics"
	"github.com/wavefronthq/wavefront-collector-for-kubernetes/plugins/sources/prometheus"

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
	metricsPrefix = "kubernetes.controlplane."
)

type provider struct {
	metrics.DefaultSourceProvider

	providers []metrics.SourceProvider
}

func NewProvider(cfg configuration.ControlPlaneSourceConfig, summaryCfg configuration.SummarySourceConfig) (metrics.SourceProvider, error) {
	var providers []metrics.SourceProvider
	for _, promCfg := range buildPromConfigs(cfg, summaryCfg) {
		provider, err := prometheus.NewPrometheusProvider(promCfg)
		if err != nil {
			return nil, fmt.Errorf("error building prometheus sources for control plane: %s", err.Error())
		}
		providers = append(providers, provider)
	}
	return &provider{providers: providers}, nil
}

func (p *provider) GetMetricsSources() []metrics.Source {
	var sources []metrics.Source
	for _, provider := range p.providers {
		sources = append(sources, provider.GetMetricsSources()...)
	}
	return sources
}

func (p *provider) Name() string {
	return metricsSource
}

func (p *provider) DiscoveryPluginConfigs() []discovery.PluginConfig {
	if !util.ScrapeAnyNodes() {
		return nil
	}
	return []discovery.PluginConfig{{
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
		Prefix: metricsPrefix,
		Filters: filter.Config{
			MetricAllowList: []string{
				metricsPrefix + "coredns.dns.request.duration.seconds.bucket",
				metricsPrefix + "coredns.dns.responses.total.counter",
			},
		},
		Collection: discovery.CollectionConfig{
			Interval: p.CollectionInterval(),
			Timeout:  p.Timeout(),
		},
		Internal: true,
	}}
}

func buildPromConfigs(cfg configuration.ControlPlaneSourceConfig, summaryCfg configuration.SummarySourceConfig) []configuration.PrometheusSourceConfig {
	var prometheusSourceConfigs []configuration.PrometheusSourceConfig

	kubeConfig, _, err := kubelet.GetKubeConfigs(summaryCfg)
	if err != nil {
		log.Infof("error %v", err)
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
		metricsPrefix + "etcd.request.duration.seconds.bucket",
		metricsPrefix + "etcd.request.duration.seconds",
		metricsPrefix + "etcd.object.counts.gauge",
		metricsPrefix + "etcd.db.total.size.in.bytes.gauge",
		metricsPrefix + "workqueue.adds.total.counter",
		metricsPrefix + "workqueue.queue.duration.seconds.bucket",
		metricsPrefix + "workqueue.queue.duration.seconds",
	}

	promSourceConfig := createPrometheusSourceConfig("etcd-workqueue", httpClientConfig, metricAllowList, nil, cfg.Collection.Interval+jitterTime)
	prometheusSourceConfigs = append(prometheusSourceConfigs, promSourceConfig)

	apiServerAllowList := []string{
		metricsPrefix + "apiserver.request.duration.seconds.bucket",
		metricsPrefix + "apiserver.request.duration.seconds",
		metricsPrefix + "apiserver.request.total.counter",
	}
	apiServerTagAllowList := map[string][]string{
		"resource": {"customresourcedefinitions", "namespaces", "lease", "nodes", "pods", "tokenreviews", "subjectaccessreviews"},
	}
	promApiServerSourceConfig := createPrometheusSourceConfig("apiserver", httpClientConfig, apiServerAllowList, apiServerTagAllowList, cfg.Collection.Interval)
	prometheusSourceConfigs = append(prometheusSourceConfigs, promApiServerSourceConfig)

	return prometheusSourceConfigs
}

func createPrometheusSourceConfig(name string, httpClientConfig httputil.ClientConfig, metricAllowList []string,
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
		Name:              name,
		UseLeaderElection: true,
	}
	return sourceConfig
}
