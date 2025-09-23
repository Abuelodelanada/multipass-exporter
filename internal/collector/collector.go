package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// MultipassInfoOutput mirrors JSON from `multipass info --format=json`
type MultipassInfoOutput struct {
	Name         string                 `json:"name"`
	State        string                 `json:"state"`
	IPv4         []string               `json:"ipv4"`
	Release      string                 `json:"release"`
	ImageHash    string                 `json:"image_hash"`
	ImageRelease string                 `json:"image_release"`
	Load         []float64              `json:"load"`
	CPUCount     string                 `json:"cpu_count"`
	Memory       MemoryInfo             `json:"memory"`
	Disks        map[string]DiskInfo    `json:"disks"`
	Mounts       map[string]interface{} `json:"mounts"`
}

type MemoryInfo struct {
	Total int64 `json:"total"`
	Used  int64 `json:"used"`
}

type DiskInfo struct {
	Total string `json:"total"`
	Used  string `json:"used"`
}

type Mount struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	SourceType  string   `json:"source_type"`
	UidMappings []UIDMap `json:"uid_mappings"`
	GidMappings []GIDMap `json:"gid_mappings"`
}

type UIDMap struct {
	HostUID     int `json:"host_uid"`
	InstanceUID int `json:"instance_uid"`
}

type GIDMap struct {
	HostGID     int `json:"host_gid"`
	InstanceGID int `json:"instance_gid"`
}

type MultipassInfoResponse struct {
	Info map[string]MultipassInfoOutput `json:"info"`
}

// CommandExecutor interface for executing commands (useful for testing)
type CommandExecutor interface {
	CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd
}

// RealCommandExecutor implements CommandExecutor using os/exec
type RealCommandExecutor struct{}

func (r RealCommandExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// MultipassCollector implements Prometheus collector
type MultipassCollector struct {
	instanceTotal       *prometheus.Desc
	instanceRunning     *prometheus.Desc
	instanceStopped     *prometheus.Desc
	instanceDeleted     *prometheus.Desc
	instanceSuspended   *prometheus.Desc
	instanceMemoryBytes *prometheus.Desc
	timeout             time.Duration
	executor            CommandExecutor
	logger              *logrus.Logger
}

func NewMultipassCollector(timeoutSeconds int) *MultipassCollector {
	return NewMultipassCollectorWithExecutor(timeoutSeconds, RealCommandExecutor{})
}

func NewMultipassCollectorWithExecutor(timeoutSeconds int, executor CommandExecutor) *MultipassCollector {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		PadLevelText:           true,
	})
	logger.SetLevel(logrus.InfoLevel)

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
		instanceDeleted: prometheus.NewDesc(
			"multipass_instances_deleted",
			"Total number of Multipass deleted instances",
			nil, nil,
		),
		instanceSuspended: prometheus.NewDesc(
			"multipass_instances_suspended",
			"Total number of Multipass suspended instances",
			nil, nil,
		),
		instanceMemoryBytes: prometheus.NewDesc(
			"multipass_instance_memory_bytes",
			"Memory usage of Multipass instances in bytes",
			[]string{"name", "release"}, nil,
		),
		timeout:  time.Duration(timeoutSeconds) * time.Second,
		executor: executor,
		logger:   logger,
	}
}

// SetLogLevel allows configuring the log level
func (c *MultipassCollector) SetLogLevel(level string) error {
	logrusLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	c.logger.SetLevel(logrusLevel)
	return nil
}

// Describe sends metrics descriptions
func (c *MultipassCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.instanceTotal
	ch <- c.instanceRunning
	ch <- c.instanceStopped
	ch <- c.instanceDeleted
	ch <- c.instanceSuspended
	ch <- c.instanceMemoryBytes
}

// Collect fetches instance count and sends to Prometheus
func (c *MultipassCollector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Info("Starting metrics collection")

	// Get multipass info once and reuse it
	data, err := c.multipassInfo()
	if err != nil {
		c.logger.WithError(err).Error("Failed to get multipass info")
		c.collectError(ch, err)
		return
	}

	// Use cached data for all metrics
	if err := c.collectInstanceTotalWithData(ch, data); err != nil {
		c.logger.WithError(err).Error("Failed to collect instance total")
		c.collectError(ch, err)
		return
	}

	if err := c.collectInstanceRunningWithData(ch, data); err != nil {
		c.logger.WithError(err).Error("Failed to collect instance running")
		c.collectError(ch, err)
		return
	}

	if err := c.collectInstanceStoppedWithData(ch, data); err != nil {
		c.logger.WithError(err).Error("Failed to collect instance stopped")
		c.collectError(ch, err)
		return
	}
	if err := c.collectInstanceDeletedWithData(ch, data); err != nil {
		c.logger.WithError(err).Error("Failed to collect instance deleted")
		c.collectError(ch, err)
		return
	}
	if err := c.collectInstanceSuspendedWithData(ch, data); err != nil {
		c.logger.WithError(err).Error("Failed to collect instance suspended")
		c.collectError(ch, err)
		return
	}
	if err := c.collectInstanceMemoryBytesWithData(ch, data); err != nil {
		c.logger.WithError(err).Error("Failed to collect instance memory bytes")
		c.collectError(ch, err)
		return
	}
}

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

func (c *MultipassCollector) collectInstanceTotalWithData(ch chan<- prometheus.Metric, data MultipassInfoResponse) error {
	count := len(data.Info)
	c.logger.WithField("count", count).Debug("Collecting instance total")

	ch <- prometheus.MustNewConstMetric(
		c.instanceTotal,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil
}

func (c *MultipassCollector) collectInstanceRunning(ch chan<- prometheus.Metric) error {
	count, err := c.getInstanceCountByState("Running")
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

func (c *MultipassCollector) collectInstanceRunningWithData(ch chan<- prometheus.Metric, data MultipassInfoResponse) error {
	count := c.getInstanceCountByStateWithData(data, "Running")
	c.logger.WithField("count", count).Debug("Collecting instance running")

	ch <- prometheus.MustNewConstMetric(
		c.instanceRunning,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil
}

func (c *MultipassCollector) collectInstanceStopped(ch chan<- prometheus.Metric) error {
	count, err := c.getInstanceCountByState("Stopped")
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

func (c *MultipassCollector) collectInstanceStoppedWithData(ch chan<- prometheus.Metric, data MultipassInfoResponse) error {
	count := c.getInstanceCountByStateWithData(data, "Stopped")
	c.logger.WithField("count", count).Debug("Collecting instance stopped")

	ch <- prometheus.MustNewConstMetric(
		c.instanceStopped,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil
}

func (c *MultipassCollector) collectInstanceDeleted(ch chan<- prometheus.Metric) error {
	count, err := c.getInstanceCountByState("Deleted")
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.instanceDeleted,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil
}

func (c *MultipassCollector) collectInstanceDeletedWithData(ch chan<- prometheus.Metric, data MultipassInfoResponse) error {
	count := c.getInstanceCountByStateWithData(data, "Deleted")
	c.logger.WithField("count", count).Debug("Collecting instance deleted")

	ch <- prometheus.MustNewConstMetric(
		c.instanceDeleted,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil
}

func (c *MultipassCollector) collectInstanceSuspended(ch chan<- prometheus.Metric) error {
	count, err := c.getInstanceCountByState("Suspended")
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.instanceSuspended,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil
}

func (c *MultipassCollector) collectInstanceSuspendedWithData(ch chan<- prometheus.Metric, data MultipassInfoResponse) error {
	count := c.getInstanceCountByStateWithData(data, "Suspended")
	c.logger.WithField("count", count).Debug("Collecting instance suspended")

	ch <- prometheus.MustNewConstMetric(
		c.instanceSuspended,
		prometheus.GaugeValue,
		float64(count),
	)
	return nil
}

func (c *MultipassCollector) collectInstanceMemoryBytesWithData(ch chan<- prometheus.Metric, data MultipassInfoResponse) error {
	c.logger.WithField("instance_count", len(data.Info)).Info("Collecting memory metrics")
	metricsCollected := 0

	for name, info := range data.Info {
		if info.Memory.Used == 0 {
			c.logger.WithField("instance", name).Debug("Skipping instance - memory usage is 0")
			continue
		}

		c.logger.WithFields(logrus.Fields{
			"instance":     name,
			"memory_bytes": info.Memory.Used,
			"release":      info.Release,
		}).Debug("Adding memory metric")
		ch <- prometheus.MustNewConstMetric(
			c.instanceMemoryBytes,
			prometheus.GaugeValue,
			float64(info.Memory.Used),
			name, info.Release,
		)
		metricsCollected++
	}

	c.logger.WithField("metrics_collected", metricsCollected).Info("Successfully collected memory metrics")
	return nil
}

func (c *MultipassCollector) collectError(ch chan<- prometheus.Metric, err error) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("multipass_error", "Error collecting metrics from Multipass", nil, nil),
		prometheus.GaugeValue, 1,
	)
}

func (c *MultipassCollector) multipassInfo() (MultipassInfoResponse, error) {
	c.logger.Debug("Executing multipass info command")
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	cmd := c.executor.CommandContext(ctx, "multipass", "info", "--format=json")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.logger.WithField("timeout", c.timeout).Error("multipass info command timed out")
			return MultipassInfoResponse{}, fmt.Errorf("multipass info timed out after %v", c.timeout)
		}
		c.logger.WithError(err).WithField("stderr", stderr.String()).Error("multipass info command failed")
		return MultipassInfoResponse{}, fmt.Errorf("multipass info failed: %w: %s", err, stderr.String())
	}

	var data MultipassInfoResponse

	if err := json.Unmarshal(out.Bytes(), &data); err != nil {
		c.logger.WithError(err).Error("Failed to parse multipass info JSON")
		return MultipassInfoResponse{}, fmt.Errorf("error parsing JSON: %w; stdout=%s; stderr=%s", err, out.String(), stderr.String())
	}

	c.logger.WithField("instance_count", len(data.Info)).Info("Successfully parsed multipass info")
	return data, nil
}

func (c *MultipassCollector) getInstanceCount() (int, error) {
	data, err := c.multipassInfo()
	if err != nil {
		return 0, err
	}
	return len(data.Info), nil
}

func (c *MultipassCollector) getInstanceCountByState(state string) (int, error) {
	data, err := c.multipassInfo()
	if err != nil {
		return 0, err
	}
	instanceCount := 0
	for _, instance := range data.Info {
		if instance.State == state {
			instanceCount++
		}
	}

	return instanceCount, nil
}

func (c *MultipassCollector) getInstanceCountByStateWithData(data MultipassInfoResponse, state string) int {
	instanceCount := 0
	for _, instance := range data.Info {
		if instance.State == state {
			instanceCount++
		}
	}
	return instanceCount
}
