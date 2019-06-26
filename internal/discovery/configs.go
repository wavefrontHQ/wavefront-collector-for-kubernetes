package discovery

import (
	"time"

	"github.com/wavefronthq/wavefront-kubernetes-collector/internal/filter"
)

// configuration for auto discovery
type Config struct {
	Global        GlobalConfig   `yaml:"global"`
	PluginConfigs []PluginConfig `yaml:"plugin_configs"`
}

// Describes global rules that define the default discovery behavior
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

	Filters filter.Config `yaml:"filters"`
}

// Describes selectors for identifying kubernetes resources
type Selectors struct {
	// The resource type the rule applies to. One of <pod|service|apiserver>. Defaults to pod.
	ResourceType string `yaml:"resourceType"`

	// the container images to match against specified as a list of glob pattern strings. Ex: 'redis*'
	Images []string `yaml:"images"`

	// map of labels to select resources by
	Labels map[string][]string `yaml:"labels"`

	// the optional namespaces to filter resources by.
	Namespaces []string `yaml:"namespaces"`
}
