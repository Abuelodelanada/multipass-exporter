package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Abuelodelanada/multipass-exporter/internal/collector"
)

func main() {
	registry := prometheus.NewRegistry()

	// Register Multipass collector
	registry.MustRegister(collector.NewMultipassCollector())

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	log.Println("Multipass Exporter is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
