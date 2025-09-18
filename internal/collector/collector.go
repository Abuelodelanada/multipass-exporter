package collector

import (
	"bytes"
	"context"
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
	instanceTotal   *prometheus.Desc
	instanceRunning *prometheus.Desc
	timeout         time.Duration
}

// NewMultipassCollector creates new collector
func NewMultipassCollector(timeoutSecond int) *MultipassCollector {
	return &MultipassCollector{
		instanceTotal: prometheus.NewDesc(
			"multipass_instance_total",
			"Total number of Multipass instances",
			nil, nil,
		),
		instanceRunning: prometheus.NewDesc(
			"multipass_instance_running",
			"Total number of Multipass running instances",
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

// getInstanceCount runs 'multipass list --format=json' with a context timeout,
// captures stdout+stderr, and parses the "list" field from the JSON output.
func (c *MultipassCollector) getInstanceCount() (int, error) {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    // Use CommandContext so the child process is killed when ctx is done.
    cmd := exec.CommandContext(ctx, "multipass", "list", "--format=json")

    // Capture both stdout and stderr for better diagnostics.
    var out bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        // If context timed out, return specific error
        if ctx.Err() == context.DeadlineExceeded {
            return 0, fmt.Errorf("multipass list timed out after %v", c.timeout)
        }
        // include stderr for debugging
        return 0, fmt.Errorf("multipass list failed: %w: %s", err, stderr.String())
    }

    var data struct {
        List []MultipassListOutput `json:"list"`
    }
    if err := json.Unmarshal(out.Bytes(), &data); err != nil {
        return 0, fmt.Errorf("error parsing JSON: %w; stdout=%s; stderr=%s", err, out.String(), stderr.String())
    }

    return len(data.List), nil
}
