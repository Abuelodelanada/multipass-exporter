package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// MultipassListOutput represents a single instance from `multipass list --format json`
type MultipassListOutput struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

// MultipassListResponse represents the root object returned by `multipass list --format json`
type MultipassListResponse struct {
	List []MultipassListOutput `json:"list"`
}

// MultipassCollector collects metrics about Multipass instances
type MultipassCollector struct {
	instanceTotal *prometheus.Desc
}

// NewMultipassCollector creates a new MultipassCollector
func NewMultipassCollector() *MultipassCollector {
	return &MultipassCollector{
		instanceTotal: prometheus.NewDesc(
			"multipass_instance_total",
			"Total number of Multipass instances",
			nil, nil,
		),
	}
}

// Describe sends the descriptors of each metric over to the provided channel.
func (c *MultipassCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.instanceTotal
}

// Collect fetches the stats and delivers them as Prometheus metrics
func (c *MultipassCollector) Collect(ch chan<- prometheus.Metric) {
	count, err := getInstanceCount()
	if err != nil {
		fmt.Printf("Error fetching instance count: %v\n", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.instanceTotal,
		prometheus.GaugeValue,
		float64(count),
	)
}

// getInstanceCount returns the number of Multipass instances currently running
func getInstanceCount() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "multipass", "list", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error executing multipass list: %w", err)
	}

	var response MultipassListResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return 0, fmt.Errorf("error parsing JSON: %w", err)
	}

	return len(response.List), nil
}
