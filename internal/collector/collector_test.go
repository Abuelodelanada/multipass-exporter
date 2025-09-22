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

func TestMultipassList_Success(t *testing.T) {
	// Test the JSON parsing logic directly
	testJSON := `{
		"list": [
			{
				"name": "instance1",
				"state": "Running",
				"ipv4": ["192.168.64.2"],
				"release": "22.04 LTS"
			},
			{
				"name": "instance2",
				"state": "Stopped",
				"ipv4": [],
				"release": "20.04 LTS"
			}
		]
	}`

	var data MultipassListResponse
	err := json.Unmarshal([]byte(testJSON), &data)

	if err != nil {
		t.Fatalf("Failed to parse test JSON: %v", err)
	}

	if len(data.List) != 2 {
		t.Errorf("Expected 2 instances, got %d", len(data.List))
	}

	if data.List[0].Name != "instance1" {
		t.Errorf("Expected first instance name 'instance1', got '%s'", data.List[0].Name)
	}

	if data.List[0].State != "Running" {
		t.Errorf("Expected first instance state 'Running', got '%s'", data.List[0].State)
	}

	if data.List[1].Name != "instance2" {
		t.Errorf("Expected second instance name 'instance2', got '%s'", data.List[1].Name)
	}

	if data.List[1].State != "Stopped" {
		t.Errorf("Expected second instance state 'Stopped', got '%s'", data.List[1].State)
	}
}

func TestCollectInstanceTotal_WithMock(t *testing.T) {
	mockJSON := `{"list": [{"name": "test1", "state": "Running", "ipv4": [], "release": "22.04 LTS"}]}`
	mockExecutor := &MockCommandExecutor{output: mockJSON}

	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)
	ch := make(chan prometheus.Metric, 1)

	err := collector.collectInstanceTotal(ch)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	select {
	case metric := <-ch:
		// Verify the metric has correct properties
		pb := &dto.Metric{}
		if err := metric.Write(pb); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}

		if *pb.Gauge.Value != 1 {
			t.Errorf("Expected metric value 1, got %f", *pb.Gauge.Value)
		}
	default:
		t.Fatal("Expected metric to be sent to channel")
	}
}

func TestCollectInstanceRunning_WithMock(t *testing.T) {
	mockJSON := `{
		"list": [
			{"name": "test1", "state": "Running", "ipv4": [], "release": "22.04 LTS"},
			{"name": "test2", "state": "Stopped", "ipv4": [], "release": "22.04 LTS"},
			{"name": "test3", "state": "Running", "ipv4": [], "release": "22.04 LTS"}
		]
	}`
	mockExecutor := &MockCommandExecutor{output: mockJSON}

	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)
	ch := make(chan prometheus.Metric, 1)

	err := collector.collectInstanceRunning(ch)

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

func TestCollectInstanceStopped_WithMock(t *testing.T) {
	mockJSON := `{
		"list": [
			{"name": "test1", "state": "Running", "ipv4": [], "release": "22.04 LTS"},
			{"name": "test2", "state": "Stopped", "ipv4": [], "release": "22.04 LTS"},
			{"name": "test3", "state": "Stopped", "ipv4": [], "release": "22.04 LTS"}
		]
	}`
	mockExecutor := &MockCommandExecutor{output: mockJSON}

	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)
	ch := make(chan prometheus.Metric, 1)

	err := collector.collectInstanceStopped(ch)

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

func TestCollectInstanceTotal_WithError(t *testing.T) {
	mockExecutor := &MockCommandExecutor{err: fmt.Errorf("command failed")}

	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)
	ch := make(chan prometheus.Metric, 1)

	err := collector.collectInstanceTotal(ch)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestCollectInstanceTotal_WithTimeout(t *testing.T) {
	mockExecutor := &FailingCommandExecutor{failWithTimeout: true}

	collector := NewMultipassCollectorWithExecutor(5, mockExecutor)
	ch := make(chan prometheus.Metric, 1)

	err := collector.collectInstanceTotal(ch)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// The error should contain some indication of failure
	if err.Error() == "" {
		t.Errorf("Expected error message, got empty error")
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

func TestMultipassListOutput_JSONUnmarshal(t *testing.T) {
	jsonStr := `{
		"name": "test-instance",
		"state": "Running",
		"ipv4": ["192.168.64.2", "10.0.0.1"],
		"release": "22.04 LTS"
	}`

	var output MultipassListOutput
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
	ch := make(chan prometheus.Metric, 10)

	err := collector.collectInstanceMemoryBytes(ch)
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

	err := collector.collectInstanceMemoryBytes(ch)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(s)-len(substr)+len(substr)] == substr
}
