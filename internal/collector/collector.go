package collector

import (
	"encoding/json"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
)

// multipassListOutput represents the JSON structure returned by `multipass list --format json`.
type multipassListOutput struct {
	List []struct {
		Name   string `json:"name"`
		State  string `json:"state"`
		IPv4   []string `json:"ipv4"`
		Release string `json:"release"`
	} `json:"list"`
}

type MultipassCollector struct {
	instanceTotal *prometheus.Desc
}

func NewMultipassCollector() *MultipassCollector {
	return &MultipassCollector{
		instanceTotal: prometheus.NewDesc(
			"multipass_instances_total",
			"Total number of Multipass instances detected",
			nil, nil,
		),
	}
}

// Describe sends the descriptors of each metric to Prometheus.
func (c *MultipassCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.instanceTotal
}

// Collect fetches the data and sends the metrics to Prometheus.
func (c *MultipassCollector) Collect(ch chan<- prometheus.Metric) {
	count, err := getInstanceCount()
	if err != nil {
		// If there is an error, export -1 so it's visible in metrics
		ch <- prometheus.MustNewConstMetric(
			c.instanceTotal,
			prometheus.GaugeValue,
			-1,
		)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.instanceTotal,
		prometheus.GaugeValue,
		float64(count),
	)
}

// getInstanceCount executes `multipass list --format json` and returns the number of instances.
func getInstanceCount() (int, error) {
	cmd := exec.Command("multipass", "list", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var data multipassListOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return 0, err
	}

	return len(data.List), nil
}
