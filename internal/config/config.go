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
	LogLevel       string `yaml:"log_level"`
}

// LoadConfig loads YAML file or returns defaults
// Returns a boolean indicating if the file was actually loaded
func LoadConfig(path string) (*Config, bool, error) {
	cfg := &Config{
		Port:           1986,
		MetricsPath:    "/metrics",
		TimeoutSeconds: 5,
		LogLevel:       "info",
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		// File missing? Use defaults
		return cfg, false, nil
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, false, fmt.Errorf("error parsing YAML: %w", err)
	}

	return cfg, true, nil
}
