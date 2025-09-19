package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Config holds exporter settings
type Config struct {
	Port           int    `yaml:"port"`
	MetricsPath    string `yaml:"metrics_path"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

// LoadConfig loads YAML file or returns defaults
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		Port:           8080,
		MetricsPath:    "/metrics",
		TimeoutSeconds: 5,
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		// File missing? Use defaults
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	return cfg, nil
}
