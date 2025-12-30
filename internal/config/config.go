package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3" //nolint:typecheck
)

// Config holds exporter settings
type Config struct {
	Port           int    `yaml:"port"`
	MetricsPath    string `yaml:"metrics_path"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	LogLevel       string `yaml:"log_level"`
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		Port:           1986,
		MetricsPath:    "/metrics",
		TimeoutSeconds: 5,
		LogLevel:       "info",
	}
}

// Validate checks if the configuration values are valid
func (c *Config) Validate() error {
	if c.Port <= 0 {
		return fmt.Errorf("port must be a positive integer, got %d", c.Port)
	}

	if c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}

	if c.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be a positive integer, got %d", c.TimeoutSeconds)
	}

	return nil
}

// LoadConfig loads YAML file or returns defaults
// Returns a boolean indicating if the file was actually loaded
func LoadConfig(path string) (*Config, bool, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		// File missing? Use defaults
		return cfg, false, nil
	}

	if err := yaml.Unmarshal(data, cfg); err != nil { //nolint:typecheck
		return nil, false, fmt.Errorf("error parsing YAML: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, false, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, true, nil
}
