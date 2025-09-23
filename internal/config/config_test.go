package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	nonExistentPath := "/tmp/nonexistent_config.yaml"
	cfg, loaded, err := LoadConfig(nonExistentPath)

	if err != nil {
		t.Fatalf("Expected no error for non-existent config file, got %v", err)
	}

	if loaded {
		t.Error("Expected loaded to be false for non-existent file")
	}

	if cfg.Port != 1986 {
		t.Errorf("Expected default port 8080, got %d", cfg.Port)
	}

	if cfg.MetricsPath != "/metrics" {
		t.Errorf("Expected default metrics path /metrics, got %s", cfg.MetricsPath)
	}

	if cfg.TimeoutSeconds != 5 {
		t.Errorf("Expected default timeout 5 seconds, got %d", cfg.TimeoutSeconds)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	configContent := `
port: 9090
metrics_path: /custom-metrics
timeout_seconds: 10
`

	tempFile := filepath.Join(t.TempDir(), "test_config.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, loaded, err := LoadConfig(tempFile)
	if err != nil {
		t.Fatalf("Expected no error for valid config file, got %v", err)
	}

	if !loaded {
		t.Error("Expected loaded to be true for existing file")
	}

	if cfg.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Port)
	}

	if cfg.MetricsPath != "/custom-metrics" {
		t.Errorf("Expected metrics path /custom-metrics, got %s", cfg.MetricsPath)
	}

	if cfg.TimeoutSeconds != 10 {
		t.Errorf("Expected timeout 10 seconds, got %d", cfg.TimeoutSeconds)
	}
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	configContent := `
port: 3000
`

	tempFile := filepath.Join(t.TempDir(), "partial_config.yaml")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, loaded, err := LoadConfig(tempFile)
	if err != nil {
		t.Fatalf("Expected no error for partial config file, got %v", err)
	}

	if !loaded {
		t.Error("Expected loaded to be true for existing file")
	}

	if cfg.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", cfg.Port)
	}

	if cfg.MetricsPath != "/metrics" {
		t.Errorf("Expected default metrics path /metrics, got %s", cfg.MetricsPath)
	}

	if cfg.TimeoutSeconds != 5 {
		t.Errorf("Expected default timeout 5 seconds, got %d", cfg.TimeoutSeconds)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	invalidYAML := `
port: 9090
metrics_path: /metrics
timeout_seconds: 10
invalid: yaml: content: [unclosed bracket
`

	tempFile := filepath.Join(t.TempDir(), "invalid_config.yaml")
	err := os.WriteFile(tempFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, loaded, err := LoadConfig(tempFile)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}

	if loaded {
		t.Error("Expected loaded to be false for invalid YAML")
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "empty_config.yaml")
	err := os.WriteFile(tempFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, loaded, err := LoadConfig(tempFile)
	if err != nil {
		t.Fatalf("Expected no error for empty config file, got %v", err)
	}

	if !loaded {
		t.Error("Expected loaded to be true for empty file (file exists)")
	}

	if cfg.Port != 1986 {
		t.Errorf("Expected default port 1986 for empty file, got %d", cfg.Port)
	}

	if cfg.MetricsPath != "/metrics" {
		t.Errorf("Expected default metrics path /metrics for empty file, got %s", cfg.MetricsPath)
	}

	if cfg.TimeoutSeconds != 5 {
		t.Errorf("Expected default timeout 5 seconds for empty file, got %d", cfg.TimeoutSeconds)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Port != 1986 {
		t.Errorf("Expected default port 1986, got %d", cfg.Port)
	}

	if cfg.MetricsPath != "/metrics" {
		t.Errorf("Expected default metrics path /metrics, got %s", cfg.MetricsPath)
	}

	if cfg.TimeoutSeconds != 5 {
		t.Errorf("Expected default timeout 5 seconds, got %d", cfg.TimeoutSeconds)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected default log level info, got %s", cfg.LogLevel)
	}
}
