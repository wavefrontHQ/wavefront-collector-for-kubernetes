package discovery

import "time"

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

// Describes rules for auto discovering pods and configuring relevant prometheus sources for metrics collection.
type PrometheusConfig struct {
	// name of the rule
	Name string `yaml:"name"`

	// map of labels to select pods by
	Labels map[string]string `yaml:"labels"`

	// the optional namespace to filter pods by.
	Namespace string `yaml:"namespace"`

	// the port to scrape for prometheus metrics. If omitted, defaults to a port-free target.
	Port string `yaml:"port"`

	// Optional. Defaults to "/metrics".
	Path string `yaml:"path"`

	// the scheme to use. defaults to "http".
	Scheme string `yaml:"scheme"`

	// prefix for metrics collected using this rule.
	Prefix string `yaml:"prefix"`

	// optional map of custom tags to include with the reported metrics.
	Tags map[string]string `yaml:"tags"`

	// optional source for metrics collected using this rule. defaults to "prom_source".
	Source string `yaml:"source"`

	// whether to include pod labels with the reported metrics. defaults to "true".
	IncludeLabels string `yaml:"includeLabels"`
}
