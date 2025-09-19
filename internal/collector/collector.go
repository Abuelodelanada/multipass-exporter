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

type MultipassListResponse struct {
	List []MultipassListOutput `json:"list"`
}


// MultipassCollector implements Prometheus collector
type MultipassCollector struct {
	instanceTotal   *prometheus.Desc
	instanceRunning *prometheus.Desc
	instanceStopped *prometheus.Desc
	timeout         time.Duration
}

// NewMultipassCollector creates new collector
func NewMultipassCollector(timeoutSeconds int) *MultipassCollector {
	return &MultipassCollector{
		instanceTotal: prometheus.NewDesc(
			"multipass_instances_total",
			"Total number of Multipass instances",
			nil, nil,
		),
		instanceRunning: prometheus.NewDesc(
			"multipass_instances_running",
			"Total number of Multipass running instances",
			nil, nil,
		),
		instanceStopped: prometheus.NewDesc(
			"multipass_instances_stopped",
			"Total number of Multipass stopped instances",
			nil, nil,
		),
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}
}

// Describe sends metrics descriptions
func (c *MultipassCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.instanceTotal
	ch <- c.instanceRunning
	ch <- c.instanceStopped
}

// Collect fetches instance count and sends to Prometheus
func (c *MultipassCollector) Collect(ch chan<- prometheus.Metric) {
	if err := c.collectInstanceTotal(ch); err != nil {
		c.collectError(ch, err)
		return
	}

	if err := c.collectInstanceRunning(ch); err != nil {
		c.collectError(ch, err)
		return
	}

	if err := c.collectInstanceStopped(ch); err != nil {
		c.collectError(ch, err)
		return
	}
}

// collectInstanceTotal sends total instance count metric
func (c *MultipassCollector) collectInstanceTotal(ch chan<- prometheus.Metric) error {
	count, err := c.getInstanceCount()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.instanceTotal,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil
}

// collectInstanceRunning sends running instance count metric
func (c *MultipassCollector) collectInstanceRunning(ch chan<- prometheus.Metric) error {
	count, err := c.getRunningInstanceCount()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.instanceRunning,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil
}

func (c *MultipassCollector) collectInstanceStopped(ch chan<- prometheus.Metric) error {
	count, err := c.getStoppedInstanceCount()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.instanceStopped,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil
}


// collectError sends error metric when something fails
func (c *MultipassCollector) collectError(ch chan<- prometheus.Metric, err error) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("multipass_error", "Error collecting metrics from Multipass", nil, nil),
		prometheus.GaugeValue, 1,
	)
}


func (c *MultipassCollector) multipassList() (MultipassListResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "multipass", "list", "--format=json")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// If context timed out, return specific error
		if ctx.Err() == context.DeadlineExceeded {
			return MultipassListResponse{}, fmt.Errorf("multipass list timed out after %v", c.timeout)
		}
		// include stderr for debugging
		return MultipassListResponse{}, fmt.Errorf("multipass list failed: %w: %s", err, stderr.String())
	}

	var data MultipassListResponse

	if err := json.Unmarshal(out.Bytes(), &data); err != nil {
		return MultipassListResponse{}, fmt.Errorf("error parsing JSON: %w; stdout=%s; stderr=%s", err, out.String(), stderr.String())
	}
	return data, nil
}


func (c *MultipassCollector) getInstanceCount() (int, error) {
    data, err := c.multipassList()
    if err != nil {
	    return 0, err
    }
    return len(data.List), nil
}

func (c *MultipassCollector) getRunningInstanceCount() (int, error) {
    data, err := c.multipassList()
    if err != nil {
        return 0, err
    }
    runningCount := 0
    for _, instance := range data.List {
        if instance.State == "Running" {
            runningCount++
        }
    }

    return runningCount, nil
}

func (c *MultipassCollector) getStoppedInstanceCount() (int, error) {
	data, err := c.multipassList()
	if err != nil {
		return 0, err
	}
	stoppedCount := 0
	for _, instance := range data.List {
		if instance.State == "Stopped" {
			stoppedCount++
		}
	}

	return stoppedCount, nil
}
