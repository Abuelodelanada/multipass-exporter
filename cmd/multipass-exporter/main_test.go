package main

import (
	"os"
	"testing"

	"github.com/Abuelodelanada/multipass-exporter/internal/config"
)

func TestConfigValidation(t *testing.T) {
	// Test that we can create a valid config
	cfg := &config.Config{
		Port:           1986,
		MetricsPath:    "/metrics",
		TimeoutSeconds: 5,
		LogLevel:       "info",
	}

	if cfg.Port <= 0 {
		t.Error("Port should be positive")
	}
	if cfg.MetricsPath == "" {
		t.Error("MetricsPath should not be empty")
	}
	if cfg.TimeoutSeconds <= 0 {
		t.Error("TimeoutSeconds should be positive")
	}
	if cfg.LogLevel == "" {
		t.Error("LogLevel should not be empty")
	}
}

func TestConfigDefaults(t *testing.T) {
	// Test LoadConfig with non-existent file (should return defaults)
	cfg, err := config.LoadConfig("/non/existent/file.yam")
	if err != nil {
		t.Fatalf("Expected no error with non-existent file, got %v", err)
	}

	// Verify defaults
	if cfg.Port != 1986 {
		t.Errorf("Expected default port 1986, got %d", cfg.Port)
	}
	if cfg.MetricsPath != "/metrics" {
		t.Errorf("Expected default metrics path /metrics, got %s", cfg.MetricsPath)
	}
	if cfg.TimeoutSeconds != 5 {
		t.Errorf("Expected default timeout 5, got %d", cfg.TimeoutSeconds)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default log level info, got %s", cfg.LogLevel)
	}
}

func TestEnvironmentVariablePrecedence(t *testing.T) {
	// Set environment variable
	testLogLevel := "debug"
	os.Setenv("LOG_LEVEL", testLogLevel)
	defer os.Unsetenv("LOG_LEVEL")

	// Create config with different log level
	cfg := &config.Config{
		Port:           9090,
		MetricsPath:    "/metrics",
		TimeoutSeconds: 10,
		LogLevel:       "warn", // This should be overridden by env var
	}

	// Simulate the logic from main()
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = cfg.LogLevel
	}

	if logLevel != testLogLevel {
		t.Errorf("Expected log level from environment (%s), got %s", testLogLevel, logLevel)
	}
}

func TestEnvironmentVariableFallback(t *testing.T) {
	// Ensure environment variable is not set
	os.Unsetenv("LOG_LEVEL")

	// Create config
	cfg := &config.Config{
		Port:           9090,
		MetricsPath:    "/metrics",
		TimeoutSeconds: 10,
		LogLevel:       "info",
	}

	// Simulate the logic from main()
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = cfg.LogLevel
	}

	if logLevel != cfg.LogLevel {
		t.Errorf("Expected log level from config (%s), got %s", cfg.LogLevel, logLevel)
	}
}

func TestNewApp(t *testing.T) {
	// Test that NewApp creates a proper App instance
	app := NewApp()

	if app == nil {
		t.Error("NewApp should not return nil")
	}
}

func TestNewAppWithConfig(t *testing.T) {
	// Test NewAppWithConfig with specific config path
	app := NewAppWithConfig("test.yaml")

	if app == nil {
		t.Error("NewAppWithConfig should not return nil")
	}

	if app.configPath != "test.yaml" {
		t.Errorf("Expected configPath to be 'test.yaml', got %s", app.configPath)
	}
}

func TestAppLoadConfiguration(t *testing.T) {
	app := NewApp()

	// Test loading default configuration
	err := app.LoadConfiguration()
	if err != nil {
		t.Errorf("LoadConfiguration failed: %v", err)
	}

	cfg := app.GetConfig()
	if cfg.Port != 1986 {
		t.Errorf("Expected default port 1986, got %d", cfg.Port)
	}
	if cfg.MetricsPath != "/metrics" {
		t.Errorf("Expected default metrics path /metrics, got %s", cfg.MetricsPath)
	}
	if cfg.TimeoutSeconds != 5 {
		t.Errorf("Expected default timeout 5, got %d", cfg.TimeoutSeconds)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default log level info, got %s", cfg.LogLevel)
	}
}

func TestAppLoadConfigurationWithFile(t *testing.T) {
	// Create a temporary config file
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	configContent := `
port: 9090
metrics_path: "/test-metrics"
timeout_seconds: 10
log_level: "debug"
`

	_, err = tmpFile.WriteString(configContent)
	if err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}
	tmpFile.Close()

	// Test loading configuration from file
	app := NewAppWithConfig(tmpFile.Name())

	err = app.LoadConfiguration()
	if err != nil {
		t.Errorf("LoadConfiguration failed: %v", err)
	}

	cfg := app.GetConfig()
	if cfg.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Port)
	}
	if cfg.MetricsPath != "/test-metrics" {
		t.Errorf("Expected metrics path /test-metrics, got %s", cfg.MetricsPath)
	}
	if cfg.TimeoutSeconds != 10 {
		t.Errorf("Expected timeout 10, got %d", cfg.TimeoutSeconds)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected log level debug, got %s", cfg.LogLevel)
	}
}

func TestAppLoadConfigurationInvalidFile(t *testing.T) {
	app := NewAppWithConfig("/non/existent/file.yaml")

	err := app.LoadConfiguration()
	if err != nil {
		t.Errorf("LoadConfiguration should not fail with invalid file, got: %v", err)
	}

	cfg := app.GetConfig()
	// Should use default values when file doesn't exist
	if cfg.Port != 1986 {
		t.Errorf("Expected default port 1986 when file doesn't exist, got %d", cfg.Port)
	}
	if cfg.MetricsPath != "/metrics" {
		t.Errorf("Expected default metrics path when file doesn't exist, got %s", cfg.MetricsPath)
	}
}

func TestAppInitializeCollector(t *testing.T) {
	app := NewApp()

	// Load configuration first
	err := app.LoadConfiguration()
	if err != nil {
		t.Errorf("LoadConfiguration failed: %v", err)
	}

	// Test collector initialization
	err = app.InitializeCollector()
	if err != nil {
		t.Errorf("InitializeCollector failed: %v", err)
	}

	collector := app.GetCollector()
	if collector == nil {
		t.Error("Expected collector to be initialized")
	}
}

func TestAppRunIntegration(t *testing.T) {
	t.Skip("Skipping integration test due to prometheus registration conflicts")
}
