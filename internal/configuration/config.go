package configuration

import (
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"time"
)

// The main configuration struct that drives the Wavefront collector
type Config struct {
	// the global interval at which data is collected. Defaults to 60 seconds.
	CollectionInterval time.Duration `yaml:"collectionInterval"`

	// the timeout for sinks to export data to Wavefront. Defaults to 20 seconds.
	SinkExportDataTimeout time.Duration `yaml:"sinkExportDataTimeout"`

	// the global per-source scrape timeout
	ScrapeTimeout time.Duration `yaml:"scrapeTimeout"`

	MaxProcs int `yaml:"maxProcs"`

	// whether auto-discovery is enabled.
	EnableDiscovery bool `yaml:"enableDiscovery"`

	// A unique identifier for your Kubernetes cluster. Defaults to k8s-cluster.
	// Included as a point tag on all metrics sent to Wavefront.
	ClusterName string `yaml:"clusterName"`

	// list of Wavefront sinks. At least 1 is required.
	Sinks []*WavefrontSinkConfig `yaml:"sinks"`

	//----- sources -----
	SummaryConfig     *SummaySourceConfig       `yaml:"kubernetes_source"`
	PrometheusConfigs []*PrometheusSourceConfig `yaml:"prometheus_sources"`
	TelegrafConfigs   []*TelegrafSourceConfig   `yaml:"telegraf_sources"`
	SystemdConfig     *SystemdSourceConfig      `yaml:"systemd_source"`
	StatsConfig       *StatsSourceConfig        `yaml:"internal_stats_source"`

	DiscoveryConfigs []discovery.PluginConfig `yaml:"discovery_configs"`
}

// Configuration options for the Wavefront sink
type WavefrontSinkConfig struct {
	//  The Wavefront URL of the form https://YOUR_INSTANCE.wavefront.com. Only required for direct ingestion.
	Server string `yaml:"server"`

	// The Wavefront API token with direct data ingestion permission. Only required for direct ingestion.
	Token string `yaml:"token"`

	// The Wavefront proxy service address of the form wavefront-proxy.default.svc.cluster.local:2878.
	ProxyAddress string `yaml:"proxyAddress"`

	// The global prefix (dot suffixed) to be added for all metrics. Defaults to empty string.
	Prefix string `yaml:"prefix"`

	// Custom tags to include with metrics reported by this sink.
	Tags map[string]string `yaml:"tags"`

	// Filters to be applied prior to emitting the metrics to Wavefront.
	Filters filter.Config `yaml:"filters"`

	// If set to true, metrics are emitted to stdout instead. Defaults to false.
	TestMode bool `yaml:"testMode"`

	// cluster name pulled in from the top level property. Internal use only.
	ClusterName string `yaml:"-"`
}

// Configuration options for the Kubernetes summary source
type SummaySourceConfig struct {
	// Defaults to empty string.
	URL string `yaml:"url"`

	// Defaults to 10255. Use 10250 for the secure port.
	KubeletPort string `yaml:"kubeletPort"`

	// Defaults to false. Set to true if kubeletPort set to 10250.
	KubeletHttps string `yaml:"kubeletHttps"`

	// Defaults to "true".
	InClusterConfig string `yaml:"inClusterConfig"`

	// Defaults to "false".
	UseServiceAccount string `yaml:"useServiceAccount"`

	// Defaults to "false".
	Insecure string `yaml:"insecure"`

	// If not using inClusterConfig, this can be set to a valid kubeConfig file provided using a config map.
	Auth string `yaml:"auth"`

	// The global prefix (dot suffixed) to be added for all metrics. Defaults to "kubernetes.".
	Prefix string `yaml:"prefix"`

	// Custom tags to include with metrics reported by this sink.
	Tags map[string]string `yaml:"tags"`

	// Filters to be applied prior to emitting the metrics to Wavefront.
	Filters filter.Config `yaml:"filters"`
}

// Configuration options for a Prometheus source
type PrometheusSourceConfig struct {
	// The URL for a Prometheus metrics endpoint. Kubernetes Service URLs work across namespaces.
	URL string `yaml:"url"`

	// The prefix (dot suffixed such as prom.) to be applied to all metrics for this source. Defaults to empty string.
	Prefix string `yaml:"prefix"`

	// The source to set for the metrics from this source. Defaults to prom_source.
	Source string `yaml:"source"`

	// Custom tags to include with metrics reported by this source, of the form tag=key1:val1&tag=key2:val2.
	Tags map[string]string `yaml:"tags"`

	// Filters to be applied to metrics collected by this source
	Filters filter.Config `yaml:"filters"`

	// internal use only
	Discovered string `yaml:"-"`
	Name       string `yaml:"-"`

	//TODO: include tls configuration?
}

// Configuration options for a Telegraf source
type TelegrafSourceConfig struct {
	// the list of plugins to be enabled
	Plugins []string `yaml:"plugins"`

	// The prefix (dot suffixed) to be added for all metrics. Defaults to empty string.
	Prefix string `yaml:"prefix"`

	// Custom tags to include with metrics reported by this source.
	Tags map[string]string `yaml:"tags"`

	// Filters to be applied to metrics collected by this source
	Filters filter.Config `yaml:"filters"`

	// internal use only
	Discovered string `yaml:"-"`
	Name       string `yaml:"-"`
}

type SystemdSourceConfig struct {
	IncludeTaskMetrics bool `yaml:"taskMetrics"`

	IncludeStartTimeMetrics bool `yaml:"startTimeMetrics"`

	IncludeRestartMetrics bool `yaml:"restartMetrics"`

	UnitWhitelist []string `yaml:"unitWhitelist"`

	UnitBlacklist []string `yaml:"unitBlacklist"`

	// The prefix (dot suffixed) to be added for all metrics. Defaults to empty string.
	Prefix string `yaml:"prefix"`

	// Custom tags to include with metrics reported by this source.
	Tags map[string]string `yaml:"tags"`

	// Filters to be applied to metrics collected by this source
	Filters filter.Config `yaml:"filters"`
}

type StatsSourceConfig struct {
	// The prefix (dot suffixed) to be added for all metrics. Defaults to "kubernetes.collector.".
	Prefix string `yaml:"prefix"`

	// Custom tags to include with metrics reported by this source.
	Tags map[string]string `yaml:"tags"`

	// Filters to be applied to metrics collected by this source.
	Filters filter.Config `yaml:"filters"`
}
