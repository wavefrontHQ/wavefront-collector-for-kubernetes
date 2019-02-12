package discovery

import (
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
)

// configuration for auto discovery
type Config struct {
	Global      GlobalConfig       `yaml:"global"`
	PromConfigs []PrometheusConfig `yaml:"prom_configs"`
}

// Describes global rules that define the default discovery behavior
type GlobalConfig struct {
	// frequency of evaluating discovery rules. Defaults to 10 minutes.
	// format is [0-9]+(ms|[smhdwy])
	DiscoveryInterval time.Duration `yaml:"discovery_interval"`
}

// Describes rules for auto discovering resources and configuring relevant prometheus sources for metrics collection.
type PrometheusConfig struct {
	// name of the rule
	Name string `yaml:"name"`

	// the resource type to discover. defaults to pod.
	// one of "pod|service".
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
