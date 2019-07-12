package discovery

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func FromFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to load discovery config file: %v", err)
	}
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("unable to load discovery config file: %v", err)
	}
	return FromYAML(contents)
}

// FromYAML loads the configuration from a blob of YAML.
func FromYAML(contents []byte) (*Config, error) {
	var cfg Config
	if err := yaml.UnmarshalStrict(contents, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse discovery config: %v", err)
	}
	return &cfg, nil
}
