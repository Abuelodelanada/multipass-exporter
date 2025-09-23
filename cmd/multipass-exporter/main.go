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

// configPath is the command line argument for configuration file path
var configPath string

func main() {
	app := NewApp()
	app.Run()
}

// App represents the main application
type App struct {
	configPath string
	cfg        *config.Config
	collector  *collector.MultipassCollector
}

func NewApp() *App {
	// Only parse flags if they haven't been parsed already
	if !flag.Parsed() {
		flag.StringVar(&configPath, "config", "", "Path to configuration file (optional)")
		flag.Parse()
	}

	return &App{
		configPath: configPath,
	}
}


func (a *App) LoadConfiguration() error {
	var err error

	if a.configPath == "" {
		// Use default configuration
		a.cfg = config.DefaultConfig()
		log.Printf("Using default configuration: port=%d, metrics_path=%s, timeout_seconds=%d, log_level=%s",
			a.cfg.Port, a.cfg.MetricsPath, a.cfg.TimeoutSeconds, a.cfg.LogLevel)
	} else {
		// Load configuration from file
		var loaded bool
		a.cfg, loaded, err = config.LoadConfig(a.configPath)
		if err != nil {
			return fmt.Errorf("failed to load config from %s: %w", a.configPath, err)
		}
		if loaded {
			log.Printf("Loaded configuration from %s: port=%d, metrics_path=%s, timeout_seconds=%d, log_level=%s",
				a.configPath, a.cfg.Port, a.cfg.MetricsPath, a.cfg.TimeoutSeconds, a.cfg.LogLevel)
		} else {
			log.Printf("Configuration file %s not found, using defaults: port=%d, metrics_path=%s, timeout_seconds=%d, log_level=%s",
				a.configPath, a.cfg.Port, a.cfg.MetricsPath, a.cfg.TimeoutSeconds, a.cfg.LogLevel)
		}
	}

	return nil
}

func (a *App) InitializeCollector() error {
	a.collector = collector.NewMultipassCollector(a.cfg.TimeoutSeconds)

	if err := a.collector.SetLogLevel(a.cfg.LogLevel); err != nil {
		log.Printf("Warning: Invalid log level '%s', using info level: %v", a.cfg.LogLevel, err)
	}

	prometheus.MustRegister(a.collector)
	return nil
}

func (a *App) StartServer() error {
	addr := fmt.Sprintf(":%d", a.cfg.Port)
	http.Handle(a.cfg.MetricsPath, promhttp.Handler())

	log.Printf("Multipass Exporter is running on %s%s", addr, a.cfg.MetricsPath)
	return http.ListenAndServe(addr, nil)
}

func (a *App) Run() {
	if err := a.LoadConfiguration(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if err := a.InitializeCollector(); err != nil {
		log.Fatalf("Collector initialization error: %v", err)
	}

	if err := a.StartServer(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func (a *App) GetConfig() *config.Config {
	return a.cfg
}

func (a *App) GetCollector() *collector.MultipassCollector {
	return a.collector
}
