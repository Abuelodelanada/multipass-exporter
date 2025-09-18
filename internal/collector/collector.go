package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// MultipassListOutput mirrors JSON from `multipass list --format=json`
type MultipassListOutput struct {
	Name    string   `json:"name"`
	State   string   `json:"state"`
	IPv4    []string `json:"ipv4"`
	Release string   `json:"release"`
}

// MultipassCollector implements Prometheus collector
type MultipassCollector struct {
	instanceTotal *prometheus.Desc
	timeout       time.Duration
}

// NewMultipassCollector creates new collector
func NewMultipassCollector(timeoutSecond int) *MultipassCollector {
	return &MultipassCollector{
		instanceTotal: prometheus.NewDesc(
			"multipass_instance_total",
			"Total number of Multipass instances",
			nil, nil,
		),
		timeout: time.Duration(timeoutSecond) * time.Second,
	}
}

// Describe sends metrics descriptions
func (c *MultipassCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.instanceTotal
}

// Collect fetches instance count and sends to Prometheus
func (c *MultipassCollector) Collect(ch chan<- prometheus.Metric) {
	count, err := c.getInstanceCount()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("multipass_error", "Error parsing JSON from Multipass", nil, nil),
			prometheus.GaugeValue, 1,
		)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.instanceTotal,
		prometheus.GaugeValue,
		float64(count),
	)
}

// getInstanceCount runs `multipass list --format=json` and parses
func (c *MultipassCollector) getInstanceCount() (int, error) {
	cmd := exec.Command("multipass", "list", "--format=json")

	var out bytes.Buffer
	cmd.Stdout = &out

	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	select {
	case err := <-done:
		if err != nil {
			return 0, err
		}
	case <-time.After(c.timeout):
		return 0, fmt.Errorf("timeout after %v", c.timeout)
	}

	var data struct {
		List []MultipassListOutput `json:"list"`
	}

	if err := json.Unmarshal(out.Bytes(), &data); err != nil {
		return 0, fmt.Errorf("error parsing JSON: %w", err)
	}

	return len(data.List), nil
}
