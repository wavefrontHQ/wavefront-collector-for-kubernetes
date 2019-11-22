// Copyright 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/discovery"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/httputil"
)

// The main configuration struct that drives the Wavefront collector
type Config struct {
	// the global interval at which data is pushed. Defaults to 60 seconds.
	FlushInterval time.Duration `yaml:"flushInterval"`

	DefaultCollectionInterval time.Duration `yaml:"defaultCollectionInterval"`

	// the timeout for sinks to export data to Wavefront. Defaults to 20 seconds.
	SinkExportDataTimeout time.Duration `yaml:"sinkExportDataTimeout"`

	// whether auto-discovery is enabled.
	EnableDiscovery bool `yaml:"enableDiscovery"`

	// whether Events is enabled.
	EnableEvents bool `yaml:"enableEvents"`

	// A unique identifier for your Kubernetes cluster. Defaults to k8s-cluster.
	// Included as a point tag on all metrics sent to Wavefront.
	ClusterName string `yaml:"clusterName"`

	// list of Wavefront sinks. At least 1 is required.
	Sinks []*WavefrontSinkConfig `yaml:"sinks"`

	// list of sources. SummarySource is mandatory. Others are optional.
	Sources *SourceConfig `yaml:"sources"`

	// configuration specific to the events collection.
	EventsConfig EventsConfig `yaml:"events"`

	DiscoveryConfig discovery.Config `yaml:"discovery"`

	// Internal use only
	Daemon bool `yaml:"-"`
}

type EventsConfig struct {
	Filters EventsFilter `yaml:"filters"`
}

type EventsFilter struct {
	TagWhitelist map[string][]string `yaml:"tagWhitelist"`
	TagBlacklist map[string][]string `yaml:"tagBlacklist"`
}

// SourceConfig contains configuration for various sources
type SourceConfig struct {
	SummaryConfig     *SummaySourceConfig       `yaml:"kubernetes_source"`
	PrometheusConfigs []*PrometheusSourceConfig `yaml:"prometheus_sources"`
	TelegrafConfigs   []*TelegrafSourceConfig   `yaml:"telegraf_sources"`
	SystemdConfig     *SystemdSourceConfig      `yaml:"systemd_source"`
	StatsConfig       *StatsSourceConfig        `yaml:"internal_stats_source"`
}

// Transforms represents transformations that can be applied to metrics at sources or sinks
type Transforms struct {
	// The source to set for the metrics. Defaults to the name of the node on which the collector is running on.
	Source string `yaml:"source"`

	// The global prefix (dot suffixed) to be added for all metrics. Default prefix varies by source/sink.
	Prefix string `yaml:"prefix"`

	// Custom tags to include with metrics.
	Tags map[string]string `yaml:"tags"`

	// Filters to be applied prior to emitting the metrics to Wavefront.
	Filters filter.Config `yaml:"filters"`
}

// Configuration options for the Wavefront sink
type WavefrontSinkConfig struct {
	Transforms `yaml:",inline"`

	//  The Wavefront URL of the form https://YOUR_INSTANCE.wavefront.com. Only required for direct ingestion.
	Server string `yaml:"server"`

	// The Wavefront API token with direct data ingestion permission. Only required for direct ingestion.
	Token string `yaml:"token"`

	// The Wavefront proxy service address of the form wavefront-proxy.default.svc.cluster.local:2878.
	ProxyAddress string `yaml:"proxyAddress"`

	// If set to true, metrics are emitted to stdout instead. Defaults to false.
	TestMode bool `yaml:"testMode"`

	// Note: Properties below are for internal use only. These cannot be set via the configuration file.

	// Internal: Cluster name pulled in from the top level property.
	ClusterName string `yaml:"-"`

	// Internal: Collector version pulled in from top level. Used for the heartbeat metric.
	Version float64 `yaml:"-"`

	// Internal: The prefix used for internal stats. Used for the heartbeat metric.
	InternalStatsPrefix string `yaml:"-"`

	// Internal: Whether event collection has been enabled
	EventsEnabled bool `yaml:"-"`
}

type CollectionConfig struct {
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

// Configuration options for the Kubernetes summary source
type SummaySourceConfig struct {
	Transforms `yaml:",inline"`

	Collection CollectionConfig `yaml:"collection"`

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
}

// Configuration options for a Prometheus source
type PrometheusSourceConfig struct {
	Transforms `yaml:",inline"`

	Collection CollectionConfig `yaml:"collection"`

	// The URL for a Prometheus metrics endpoint. Kubernetes Service URLs work across namespaces.
	URL string `yaml:"url"`

	// Optional HTTP client configuration.
	HTTPClientConfig httputil.ClientConfig `yaml:"httpConfig"`

	// internal use only
	Discovered string `yaml:"-"`
	Name       string `yaml:"-"`
}

// Configuration options for a Telegraf source
type TelegrafSourceConfig struct {
	Transforms `yaml:",inline"`

	Collection CollectionConfig `yaml:"collection"`

	// the list of plugins to be enabled
	Plugins []string `yaml:"plugins"`

	// The configuration specific to a plugin provided in toml format: https://github.com/toml-lang/toml
	// parsed using https://github.com/influxdata/toml
	Conf string `yaml:"conf"`

	// internal use only
	Discovered string `yaml:"-"`
	Name       string `yaml:"-"`
}

type SystemdSourceConfig struct {
	Transforms `yaml:",inline"`

	Collection CollectionConfig `yaml:"collection"`

	IncludeTaskMetrics bool `yaml:"taskMetrics"`

	IncludeStartTimeMetrics bool `yaml:"startTimeMetrics"`

	IncludeRestartMetrics bool `yaml:"restartMetrics"`

	UnitWhitelist []string `yaml:"unitWhitelist"`

	UnitBlacklist []string `yaml:"unitBlacklist"`
}

type StatsSourceConfig struct {
	Transforms `yaml:",inline"`

	Collection CollectionConfig `yaml:"collection"`
}
