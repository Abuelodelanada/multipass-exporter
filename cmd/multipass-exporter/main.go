package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Abuelodelanada/multipass-exporter/internal/collector"
	"github.com/Abuelodelanada/multipass-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	c := collector.NewMultipassCollector(cfg.TimeoutSecond)
	prometheus.MustRegister(c)

	addr := fmt.Sprintf(":%d", cfg.Port)
	http.Handle(cfg.MetricsPath, promhttp.Handler())

	log.Printf("Multipass Exporter is running on %s%s", addr, cfg.MetricsPath)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
