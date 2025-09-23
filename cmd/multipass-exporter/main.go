package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/Abuelodelanada/multipass-exporter/internal/collector"
	"github.com/Abuelodelanada/multipass-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file (optional)")
	flag.Parse()

	var cfg *config.Config
	var err error

	if configPath == "" {
		// Use default configuration
		cfg = &config.Config{
			Port:           8080,
			MetricsPath:    "/metrics",
			TimeoutSeconds: 5,
			LogLevel:       "info",
		}
		log.Printf("Using default configuration: port=%d, metrics_path=%s, timeout_seconds=%d, log_level=%s",
			cfg.Port, cfg.MetricsPath, cfg.TimeoutSeconds, cfg.LogLevel)
	} else {
		// Load configuration from file
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			log.Fatalf("failed to load config from %s: %v", configPath, err)
		}
		log.Printf("Loaded configuration from %s: port=%d, metrics_path=%s, timeout_seconds=%d, log_level=%s",
			configPath, cfg.Port, cfg.MetricsPath, cfg.TimeoutSeconds, cfg.LogLevel)
	}

	c := collector.NewMultipassCollector(cfg.TimeoutSeconds)

	// Configure log level from config file
	if err := c.SetLogLevel(cfg.LogLevel); err != nil {
		log.Printf("Warning: Invalid log level '%s', using info level: %v", cfg.LogLevel, err)
	}

	prometheus.MustRegister(c)

	addr := fmt.Sprintf(":%d", cfg.Port)
	http.Handle(cfg.MetricsPath, promhttp.Handler())

	log.Printf("Multipass Exporter is running on %s%s", addr, cfg.MetricsPath)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
