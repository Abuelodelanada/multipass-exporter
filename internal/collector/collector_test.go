package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// MockCommandExecutor for testing
type MockCommandExecutor struct {
	output string
	err    error
}

func (m *MockCommandExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	if m.err != nil {
		// Return a command that will fail
		return exec.CommandContext(ctx, "false")
	}

	// Create a command that outputs our mock data
	cmd := exec.CommandContext(ctx, "echo", m.output)
	return cmd
}

// FailingCommandExecutor for testing error cases
type FailingCommandExecutor struct {
	failWithTimeout bool
	stderr          string
}

func (f *FailingCommandExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	if f.failWithTimeout {
		// Simulate timeout by using a very short context
		shortCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		return exec.CommandContext(shortCtx, "sleep", "1")
	}

	// Simulate command failure
	cmd := exec.CommandContext(ctx, "false")
	if f.stderr != "" {
		// We can't easily mock stderr with the current approach
		// but we can simulate failure
	}
	return cmd
}

func TestMultipassInfo_Success(t *testing.T) {
	// Test the JSON parsing logic directly
	testJSON := `{
		"info": {
			"instance1": {
				"name": "instance1",
				"state": "Running",
				"ipv4": ["192.168.64.2"],
				"release": "22.04 LTS",
				"memory": {
					"total": 1073741824,
					"used": 536870912
				}
			},
			"instance2": {
				"name": "instance2",
				"state": "Stopped",
				"ipv4": [],
				"release": "20.04 LTS",
				"memory": {
					"total": 1073741824,
					"used": 268435456
				}
			}
		}
	}`

	var data MultipassInfoResponse
	err := json.Unmarshal([]byte(testJSON), &data)

	if err != nil {
		t.Fatalf("Failed to parse test JSON: %v", err)
	}

	if len(data.Info) != 2 {
		t.Errorf("Expected 2 instances, got %d", len(data.Info))
	}

	instance1 := data.Info["instance1"]
	if instance1.Name != "instance1" {
		t.Errorf("Expected first instance name 'instance1', got '%s'", instance1.Name)
	}

	if instance1.State != "Running" {
		t.Errorf("Expected first instance state 'Running', got '%s'", instance1.State)
	}

	instance2 := data.Info["instance2"]
	if instance2.Name != "instance2" {
		t.Errorf("Expected second instance name 'instance2', got '%s'", instance2.Name)
	}

	if instance2.State != "Stopped" {
		t.Errorf("Expected second instance state 'Stopped', got '%s'", instance2.State)
	}
}






func TestCollectError(t *testing.T) {
	collector := NewMultipassCollectorWithExecutor(5, &MockCommandExecutor{})
	ch := make(chan prometheus.Metric, 1)

	collector.collectError(ch, fmt.Errorf("test error"))

	select {
	case metric := <-ch:
		pb := &dto.Metric{}
		if err := metric.Write(pb); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}

		if *pb.Gauge.Value != 1 {
			t.Errorf("Expected error metric value 1, got %f", *pb.Gauge.Value)
		}
	default:
		t.Fatal("Expected error metric to be sent to channel")
	}
}

func TestNewMultipassCollector(t *testing.T) {
	timeoutSeconds := 10
	collector := NewMultipassCollector(timeoutSeconds)

	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}

	if collector.timeout != time.Duration(timeoutSeconds)*time.Second {
		t.Errorf("Expected timeout %v, got %v",
			time.Duration(timeoutSeconds)*time.Second, collector.timeout)
	}

	if collector.instanceTotal == nil {
		t.Error("Expected instanceTotal descriptor to be set, got nil")
	}

	if collector.instanceRunning == nil {
		t.Error("Expected instanceRunning descriptor to be set, got nil")
	}

	if collector.instanceStopped == nil {
		t.Error("Expected instanceStopped descriptor to be set, got nil")
	}

	if collector.executor == nil {
		t.Error("Expected executor to be set, got nil")
	}
}

func TestNewMultipassCollectorWithExecutor(t *testing.T) {
	mockExecutor := &MockCommandExecutor{}
	timeoutSeconds := 10
	collector := NewMultipassCollectorWithExecutor(timeoutSeconds, mockExecutor)

	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}

	if collector.executor != mockExecutor {
		t.Error("Expected custom executor to be set")
	}
}

func TestDescribe(t *testing.T) {
	collector := NewMultipassCollector(5)

	ch := make(chan *prometheus.Desc, 10)

	collector.Describe(ch)

	close(ch)
	descriptions := make([]*prometheus.Desc, 0)
	for desc := range ch {
		descriptions = append(descriptions, desc)
	}

	if len(descriptions) != 6 {
		t.Errorf("Expected 6 metric descriptions, got %d", len(descriptions))
	}
}

func TestMultipassInfoOutput_JSONUnmarshal(t *testing.T) {
	jsonStr := `{
		"name": "test-instance",
		"state": "Running",
		"ipv4": ["192.168.64.2", "10.0.0.1"],
		"release": "22.04 LTS",
		"memory": {
			"total": 2147483648,
			"used": 1073741824
		}
	}`

	var output MultipassInfoOutput
	err := json.Unmarshal([]byte(jsonStr), &output)

	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if output.Name != "test-instance" {
		t.Errorf("Expected name 'test-instance', got '%s'", output.Name)
	}

	if output.State != "Running" {
		t.Errorf("Expected state 'Running', got '%s'", output.State)
	}

	if len(output.IPv4) != 2 {
		t.Errorf("Expected 2 IP addresses, got %d", len(output.IPv4))
	}

	if output.IPv4[0] != "192.168.64.2" {
		t.Errorf("Expected first IP '192.168.64.2', got '%s'", output.IPv4[0])
	}

	if output.Release != "22.04 LTS" {
		t.Errorf("Expected release '22.04 LTS', got '%s'", output.Release)
	}

	if output.Memory.Total != 2147483648 {
		t.Errorf("Expected memory total 2147483648, got %d", output.Memory.Total)
	}

	if output.Memory.Used != 1073741824 {
		t.Errorf("Expected memory used 1073741824, got %d", output.Memory.Used)
	}
}

func TestRealCommandExecutor(t *testing.T) {
	executor := RealCommandExecutor{}
	ctx := context.Background()
	cmd := executor.CommandContext(ctx, "echo", "test")

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
	}

	// Test that the command actually works
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	if string(output) != "test\n" {
		t.Errorf("Expected output 'test\\n', got '%s'", string(output))
	}
}

func TestCollectInstanceMemoryBytes_WithMock(t *testing.T) {
	mockJSON := `{
		"info": {
			"instance1": {
				"name": "instance1",
				"state": "Running",
				"ipv4": ["192.168.64.2"],
				"release": "22.04 LTS",
				"memory": {
					"total": 1610612736,
					"used": 536870912
				}
			},
			"instance2": {
				"name": "instance2",
				"state": "Stopped",
				"ipv4": [],
				"release": "20.04 LTS",
				"memory": {
					"total": 1073741824,
					"used": 268435456
				}
			}
		}
	}`
	mockExecutor := &MockCommandExecutor{output: mockJSON}

	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)

	// Parse the JSON manually to create the data object
	var data MultipassInfoResponse
	if err := json.Unmarshal([]byte(mockJSON), &data); err != nil {
		t.Fatalf("Failed to parse mock JSON: %v", err)
	}

	ch := make(chan prometheus.Metric, 10)

	err := collector.collectInstanceMemoryBytesWithData(ch, data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	close(ch)

	metricCount := 0
	var values []float64
	var names []string
	var releases []string

	for metric := range ch {
		metricCount++
		pb := &dto.Metric{}
		if err := metric.Write(pb); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}

		values = append(values, *pb.Gauge.Value)
		if pb.Label != nil {
			for _, label := range pb.Label {
				if label.GetName() == "name" {
					names = append(names, label.GetValue())
				}
				if label.GetName() == "release" {
					releases = append(releases, label.GetValue())
				}
			}
		}
	}

	if metricCount != 2 {
		t.Errorf("Expected 2 metrics, got %d", metricCount)
	}

	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}

	if values[0] != 536870912 && values[1] != 536870912 {
		t.Errorf("Expected one metric to be 536870912 (512MB), but got %f and %f", values[0], values[1])
	}
	if values[0] != 268435456 && values[1] != 268435456 {
		t.Errorf("Expected one metric to be 268435456 (256MB), but got %f and %f", values[0], values[1])
	}
}

func TestCollectInstanceMemoryBytes_WithError(t *testing.T) {
	mockExecutor := &MockCommandExecutor{err: fmt.Errorf("command failed")}

	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)
	ch := make(chan prometheus.Metric, 1)

	// Create empty data for error case
	data := MultipassInfoResponse{Info: make(map[string]MultipassInfoOutput)}

	err := collector.collectInstanceMemoryBytesWithData(ch, data)
	if err != nil {
		t.Fatalf("Expected no error with empty data, got %v", err)
	}
}

// Helper function
func TestSetLogLevel(t *testing.T) {
	collector := NewMultipassCollector(5)

	// Test valid log levels
	validLevels := []string{"debug", "info", "warn", "error", "fatal", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	for _, level := range validLevels {
		err := collector.SetLogLevel(level)
		if err != nil {
			t.Errorf("Expected no error for level '%s', got %v", level, err)
		}
	}

	// Test invalid log level
	err := collector.SetLogLevel("invalid")
	if err == nil {
		t.Error("Expected error for invalid log level, got nil")
	}
}

func TestCollectInstanceTotalWithData(t *testing.T) {
	mockJSON := `{
		"info": {
			"test1": {"name": "test1", "state": "Running", "ipv4": [], "release": "22.04 LTS", "memory": {"total": 1073741824, "used": 536870912}},
			"test2": {"name": "test2", "state": "Stopped", "ipv4": [], "release": "20.04 LTS", "memory": {"total": 1073741824, "used": 268435456}}
		}
	}`
	mockExecutor := &MockCommandExecutor{output: mockJSON}
	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)

	var data MultipassInfoResponse
	if err := json.Unmarshal([]byte(mockJSON), &data); err != nil {
		t.Fatalf("Failed to parse mock JSON: %v", err)
	}

	ch := make(chan prometheus.Metric, 1)
	err := collector.collectInstanceTotalWithData(ch, data)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	select {
	case metric := <-ch:
		pb := &dto.Metric{}
		if err := metric.Write(pb); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}
		if *pb.Gauge.Value != 2 {
			t.Errorf("Expected total count 2, got %f", *pb.Gauge.Value)
		}
	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}

func TestCollectInstanceRunningWithData(t *testing.T) {
	mockJSON := `{
		"info": {
			"test1": {"name": "test1", "state": "Running", "ipv4": [], "release": "22.04 LTS", "memory": {"total": 1073741824, "used": 536870912}},
			"test2": {"name": "test2", "state": "Stopped", "ipv4": [], "release": "20.04 LTS", "memory": {"total": 1073741824, "used": 268435456}},
			"test3": {"name": "test3", "state": "Running", "ipv4": [], "release": "22.04 LTS", "memory": {"total": 1073741824, "used": 536870912}}
		}
	}`
	mockExecutor := &MockCommandExecutor{output: mockJSON}
	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)

	var data MultipassInfoResponse
	if err := json.Unmarshal([]byte(mockJSON), &data); err != nil {
		t.Fatalf("Failed to parse mock JSON: %v", err)
	}

	ch := make(chan prometheus.Metric, 1)
	err := collector.collectInstanceRunningWithData(ch, data)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	select {
	case metric := <-ch:
		pb := &dto.Metric{}
		if err := metric.Write(pb); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}
		if *pb.Gauge.Value != 2 {
			t.Errorf("Expected running count 2, got %f", *pb.Gauge.Value)
		}
	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}

func TestCollectInstanceStoppedWithData(t *testing.T) {
	mockJSON := `{
		"info": {
			"test1": {"name": "test1", "state": "Running", "ipv4": [], "release": "22.04 LTS", "memory": {"total": 1073741824, "used": 536870912}},
			"test2": {"name": "test2", "state": "Stopped", "ipv4": [], "release": "20.04 LTS", "memory": {"total": 1073741824, "used": 268435456}},
			"test3": {"name": "test3", "state": "Stopped", "ipv4": [], "release": "22.04 LTS", "memory": {"total": 1073741824, "used": 536870912}}
		}
	}`
	mockExecutor := &MockCommandExecutor{output: mockJSON}
	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)

	var data MultipassInfoResponse
	if err := json.Unmarshal([]byte(mockJSON), &data); err != nil {
		t.Fatalf("Failed to parse mock JSON: %v", err)
	}

	ch := make(chan prometheus.Metric, 1)
	err := collector.collectInstanceStoppedWithData(ch, data)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	select {
	case metric := <-ch:
		pb := &dto.Metric{}
		if err := metric.Write(pb); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}
		if *pb.Gauge.Value != 2 {
			t.Errorf("Expected stopped count 2, got %f", *pb.Gauge.Value)
		}
	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}

func TestCollectInstanceDeletedWithData(t *testing.T) {
	mockJSON := `{
		"info": {
			"test1": {"name": "test1", "state": "Running", "ipv4": [], "release": "22.04 LTS", "memory": {"total": 1073741824, "used": 536870912}},
			"test2": {"name": "test2", "state": "Deleted", "ipv4": [], "release": "20.04 LTS", "memory": {"total": 1073741824, "used": 268435456}},
			"test3": {"name": "test3", "state": "Deleted", "ipv4": [], "release": "22.04 LTS", "memory": {"total": 1073741824, "used": 536870912}}
		}
	}`
	mockExecutor := &MockCommandExecutor{output: mockJSON}
	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)

	var data MultipassInfoResponse
	if err := json.Unmarshal([]byte(mockJSON), &data); err != nil {
		t.Fatalf("Failed to parse mock JSON: %v", err)
	}

	ch := make(chan prometheus.Metric, 1)
	err := collector.collectInstanceDeletedWithData(ch, data)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	select {
	case metric := <-ch:
		pb := &dto.Metric{}
		if err := metric.Write(pb); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}
		if *pb.Gauge.Value != 2 {
			t.Errorf("Expected deleted count 2, got %f", *pb.Gauge.Value)
		}
	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}

func TestCollectInstanceSuspendedWithData(t *testing.T) {
	mockJSON := `{
		"info": {
			"test1": {"name": "test1", "state": "Running", "ipv4": [], "release": "22.04 LTS", "memory": {"total": 1073741824, "used": 536870912}},
			"test2": {"name": "test2", "state": "Suspended", "ipv4": [], "release": "20.04 LTS", "memory": {"total": 1073741824, "used": 268435456}},
			"test3": {"name": "test3", "state": "Suspended", "ipv4": [], "release": "22.04 LTS", "memory": {"total": 1073741824, "used": 536870912}}
		}
	}`
	mockExecutor := &MockCommandExecutor{output: mockJSON}
	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)

	var data MultipassInfoResponse
	if err := json.Unmarshal([]byte(mockJSON), &data); err != nil {
		t.Fatalf("Failed to parse mock JSON: %v", err)
	}

	ch := make(chan prometheus.Metric, 1)
	err := collector.collectInstanceSuspendedWithData(ch, data)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	select {
	case metric := <-ch:
		pb := &dto.Metric{}
		if err := metric.Write(pb); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}
		if *pb.Gauge.Value != 2 {
			t.Errorf("Expected suspended count 2, got %f", *pb.Gauge.Value)
		}
	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}

func TestCollectMain(t *testing.T) {
	mockJSON := `{
		"info": {
			"test1": {"name": "test1", "state": "Running", "ipv4": [], "release": "22.04 LTS", "memory": {"total": 1073741824, "used": 536870912}},
			"test2": {"name": "test2", "state": "Stopped", "ipv4": [], "release": "20.04 LTS", "memory": {"total": 1073741824, "used": 268435456}}
		}
	}`
	mockExecutor := &MockCommandExecutor{output: mockJSON}
	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)

	close(ch)

	metricsCount := 0
	for range ch {
		metricsCount++
	}

	// Should collect: total, running, stopped, deleted, suspended, memory metrics, and potentially error
	if metricsCount < 6 {
		t.Errorf("Expected at least 6 metrics, got %d", metricsCount)
	}
}



func TestGetInstanceCountByStateWithData(t *testing.T) {
	data := MultipassInfoResponse{
		Info: map[string]MultipassInfoOutput{
			"instance1": {State: "Running"},
			"instance2": {State: "Running"},
			"instance3": {State: "Stopped"},
			"instance4": {State: "Running"},
		},
	}

	collector := NewMultipassCollector(5)

	runningCount := collector.getInstanceCountByStateWithData(data, "Running")
	stoppedCount := collector.getInstanceCountByStateWithData(data, "Stopped")
	deletedCount := collector.getInstanceCountByStateWithData(data, "Deleted")

	if runningCount != 3 {
		t.Errorf("Expected 3 running instances, got %d", runningCount)
	}
	if stoppedCount != 1 {
		t.Errorf("Expected 1 stopped instance, got %d", stoppedCount)
	}
	if deletedCount != 0 {
		t.Errorf("Expected 0 deleted instances, got %d", deletedCount)
	}
}

func TestCollectInstanceMemoryBytesWithDataEdgeCases(t *testing.T) {
	collector := NewMultipassCollector(5)

	// Test with no instances
	emptyData := MultipassInfoResponse{Info: make(map[string]MultipassInfoOutput)}
	ch := make(chan prometheus.Metric, 1)
	err := collector.collectInstanceMemoryBytesWithData(ch, emptyData)

	if err != nil {
		t.Fatalf("Expected no error with empty data, got %v", err)
	}

	// Test with instances having zero memory usage
	zeroMemoryData := MultipassInfoResponse{
		Info: map[string]MultipassInfoOutput{
			"instance1": {
				Name:  "instance1",
				State: "Running",
				Memory: MemoryInfo{
					Total: 1073741824,
					Used:  0,
				},
			},
		},
	}

	ch = make(chan prometheus.Metric, 1)
	err = collector.collectInstanceMemoryBytesWithData(ch, zeroMemoryData)

	if err != nil {
		t.Fatalf("Expected no error with zero memory usage, got %v", err)
	}

	// Verify no metrics were sent (since memory usage is 0)
	select {
	case <-ch:
		t.Fatal("Expected no metrics when memory usage is 0")
	default:
		// Expected behavior
	}
}
