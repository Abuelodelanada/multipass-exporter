package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Abuelodelanada/multipass-exporter/internal/collector"
)

func main() {
	// Create the Multipass collector
	multipassCollector := collector.NewMultipassCollector()

	// Register the collector with Prometheus
	prometheus.MustRegister(multipassCollector)

	// Expose the metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	addr := ":8080"
	log.Printf("Multipass Exporter is running on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
