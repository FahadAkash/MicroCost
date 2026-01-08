package models

import (
	"testing"
	"time"
)

func TestNewMetricsSnapshot(t *testing.T) {
	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	snapshot := NewMetricsSnapshot(start, end)

	if snapshot == nil {
		t.Fatal("NewMetricsSnapshot returned nil")
	}

	if snapshot.Services == nil {
		t.Error("Services map not initialized")
	}

	if snapshot.TimeRange.Start != start {
		t.Error("Start time not set correctly")
	}

	if snapshot.TimeRange.End != end {
		t.Error("End time not set correctly")
	}
}

func TestAddServiceMetrics(t *testing.T) {
	snapshot := NewMetricsSnapshot(time.Now().Add(-1*time.Hour), time.Now())

	serviceMetrics := &ServiceMetrics{
		ServiceName: "test-service",
		Endpoints:   make(map[string]*EndpointMetrics),
	}

	snapshot.AddServiceMetrics(serviceMetrics)

	if len(snapshot.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(snapshot.Services))
	}

	retrieved, exists := snapshot.GetServiceMetrics("test-service")
	if !exists {
		t.Error("Service metrics not found")
	}

	if retrieved.ServiceName != "test-service" {
		t.Errorf("Expected 'test-service', got '%s'", retrieved.ServiceName)
	}
}

func TestResourceMetrics(t *testing.T) {
	rm := &ResourceMetrics{
		CPUCores:     2.5,
		MemoryMB:     1024.0,
		NetworkInMB:  100.0,
		NetworkOutMB: 50.0,
		Timestamp:    time.Now(),
	}

	if rm.CPUCores != 2.5 {
		t.Errorf("Expected CPU 2.5, got %f", rm.CPUCores)
	}

	if rm.MemoryMB != 1024.0 {
		t.Errorf("Expected Memory 1024.0, got %f", rm.MemoryMB)
	}
}

func TestPerformanceMetrics(t *testing.T) {
	pm := &PerformanceMetrics{
		RequestRate: 100.0,
		ErrorRate:   0.5,
		LatencyAvg:  50 * time.Millisecond,
		LatencyP50:  40 * time.Millisecond,
		LatencyP95:  100 * time.Millisecond,
		LatencyP99:  200 * time.Millisecond,
		Timestamp:   time.Now(),
	}

	if pm.RequestRate != 100.0 {
		t.Errorf("Expected request rate 100.0, got %f", pm.RequestRate)
	}

	if pm.LatencyP95 != 100*time.Millisecond {
		t.Errorf("Expected P95 100ms, got %v", pm.LatencyP95)
	}
}

func TestEndpointMetrics(t *testing.T) {
	em := &EndpointMetrics{
		Service:  "test-service",
		Endpoint: "/api/test",
		Method:   "GET",
		Resource: &ResourceMetrics{
			CPUCores: 1.0,
			MemoryMB: 512.0,
		},
		Performance: &PerformanceMetrics{
			RequestRate: 50.0,
			LatencyAvg:  30 * time.Millisecond,
		},
	}

	if em.Service != "test-service" {
		t.Errorf("Expected service 'test-service', got '%s'", em.Service)
	}

	if em.Resource.CPUCores != 1.0 {
		t.Errorf("Expected CPU 1.0, got %f", em.Resource.CPUCores)
	}

	if em.Performance.RequestRate != 50.0 {
		t.Errorf("Expected request rate 50.0, got %f", em.Performance.RequestRate)
	}
}
