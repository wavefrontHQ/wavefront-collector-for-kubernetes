package configuration

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// FromFile loads the configuration from a given file
func FromFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to load configuration file: %v", err)
	}
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("unable to load configuration file: %v", err)
	}
	return FromYAML(contents)
}

// FromYAML loads the configuration from a blob of YAML.
func FromYAML(contents []byte) (*Config, error) {
	var cfg Config
	if err := yaml.UnmarshalStrict(contents, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse configuration: %v", err)
	}
	return &cfg, nil
}
