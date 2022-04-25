// Copyright 2018-2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"time"

	"github.com/wavefronthq/wavefront-collector-for-kubernetes/internal/filter"
)

// configuration for auto discovery
type Config struct {
	// frequency of evaluating discovery rules. Defaults to 5 minutes.
	// format is [0-9]+(ms|[smhdwy])
	DiscoveryInterval time.Duration `yaml:"discovery_interval"`

	// the annotation prefix for annotations based discovery
	// if specified, this substitutes the default prefix such as "prometheus.io" etc that is used during discovery
	AnnotationPrefix string `yaml:"annotation_prefix"`

	// AnnotationExcludes represents Selector rules for excluding kubernetes resources from annotation based discovery
	AnnotationExcludes []Selectors `yaml:"annotation_excludes"`

	// enables support for sourcing plugin configurations dynamically from configmaps
	// with the annotation "wavefront.com/discovery-config: 'true'". Defaults to false.
	EnableRuntimePlugins bool `yaml:"enable_runtime_plugins"`

	// disables annotation based discovery. Defaults to false.
	DisableAnnotationDiscovery bool `yaml:"disable_annotation_discovery"`

	// list of discovery rules
	PluginConfigs []PluginConfig `yaml:"plugins"`

	// Deprecated: ignored. Use top level configuration properties.
	Global GlobalConfig `yaml:"global"`

	// Deprecated: Use PluginConfigs instead.
	PromConfigs []PrometheusConfig `yaml:"prom_configs"`
}

// Describes global rules that define the default discovery behavior
// Deprecated: This configuration is ignored and retained for backwards compatibility
type GlobalConfig struct {
	// frequency of evaluating discovery rules. Defaults to 10 minutes.
	// format is [0-9]+(ms|[smhdwy])
	DiscoveryInterval time.Duration `yaml:"discovery_interval"`
}

// Describes rules for auto discovering supported services
type PluginConfig struct {
	// the unique name for this configuration rule. Used internally as map keys and needs to be unique per rule.
	Name string `yaml:"name"`

	// the plugin type, for example: 'prometheus' or 'telegraf/redis'
	Type string `yaml:"type"`

	// the selectors for identifying matching kubernetes resources
	Selectors Selectors `yaml:"selectors"`

	// the port to be monitored on the container
	Port string `yaml:"port"`

	// the scheme to use. Defaults to "http".
	Scheme string `yaml:"scheme"`

	// Optional. Defaults to "/metrics" for prometheus plugin type. Empty string for telegraf plugins.
	Path string `yaml:"path"`

	// configuration specific to the plugin.
	// For telegraf based plugins config is provided in toml format: https://github.com/toml-lang/toml
	// and parsed using https://github.com/influxdata/toml
	Conf string `yaml:"conf"`

	// Optional static source for metrics collected using this rule. Defaults to agent node name.
	Source string `yaml:"source"`

	// prefix for metrics collected using this rule. Defaults to empty string.
	Prefix string `yaml:"prefix"`

	// optional map of custom tags to include with the reported metrics
	Tags map[string]string `yaml:"tags"`

	// whether to include resource labels with the reported metrics. Defaults to "true".
	IncludeLabels string `yaml:"includeLabels"`

	Filters    filter.Config    `yaml:"filters"`
	Collection CollectionConfig `yaml:"collection"`

    // Internal: whether this plugin config was produced internally
    Internal bool `yaml:"-"`
}

type CollectionConfig struct {
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

// Describes selectors for identifying kubernetes resources
type Selectors struct {
	// The resource type the rule applies to. One of <pod|service|node>. Defaults to pod.
	ResourceType string `yaml:"resourceType"`

	// the container images to match against specified as a list of glob pattern strings. Ex: 'redis*'
	Images []string `yaml:"images"`

	// map of labels to select resources by
	Labels map[string][]string `yaml:"labels"`

	// the optional namespaces to filter resources by.
	Namespaces []string `yaml:"namespaces"`
}

// Deprecated: Use PluginConfig's instead.
type PrometheusConfig struct {
	// name of the rule
	Name string `yaml:"name"`

	// the resource type to discover. defaults to pod.
	// one of "pod|service|apiserver".
	ResourceType string `yaml:"resourceType"`

	// map of labels to select resources by
	Labels map[string]string `yaml:"labels"`

	// the optional namespace to filter resources by.
	Namespace string `yaml:"namespace"`

	// the port to scrape for prometheus metrics. If omitted, defaults to a port-free target.
	Port string `yaml:"port"`

	// Optional. Defaults to "/metrics".
	Path string `yaml:"path"`

	// the scheme to use. Defaults to "http".
	Scheme string `yaml:"scheme"`

	// prefix for metrics collected using this rule. Defaults to empty string.
	Prefix string `yaml:"prefix"`

	// optional map of custom tags to include with the reported metrics.
	Tags map[string]string `yaml:"tags"`

	// optional source for metrics collected using this rule. Defaults to the name of the Kubernetes resource.
	Source string `yaml:"source"`

	// whether to include resource labels with the reported metrics. Defaults to "true".
	IncludeLabels string `yaml:"includeLabels"`

	Filters filter.Config `yaml:"filters"`
}
