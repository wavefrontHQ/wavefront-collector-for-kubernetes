package httputil

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// FromYAML loads the configuration from a blob of YAML.
func FromYAML(contents []byte) (ClientConfig, error) {
	var cfg ClientConfig
	if err := yaml.UnmarshalStrict(contents, &cfg); err != nil {
		return ClientConfig{}, fmt.Errorf("unable to parse http configuration: %v", err)
	}
	return cfg, nil
}
