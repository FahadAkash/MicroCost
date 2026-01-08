package models

import (
	"testing"
	"time"
)

func TestNewCostReport(t *testing.T) {
	costModel := &CostModel{
		CPUCostPerCoreHour:  0.05,
		MemoryCostPerGBHour: 0.01,
		NetworkCostPerGB:    0.09,
		Provider:            "aws",
		Region:              "us-east-1",
	}

	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
	}

	report := NewCostReport(costModel, timeRange)

	if report == nil {
		t.Fatal("NewCostReport returned nil")
	}

	if report.Services == nil {
		t.Error("Services map not initialized")
	}

	if report.CostModel != costModel {
		t.Error("Cost model not set correctly")
	}

	if report.TotalCost != 0.0 {
		t.Errorf("Expected initial total cost 0.0, got %f", report.TotalCost)
	}
}

func TestAddServiceCost(t *testing.T) {
	report := NewCostReport(&CostModel{Provider: "aws"}, TimeRange{})

	serviceCost := &ServiceCost{
		ServiceName: "test-service",
		TotalCost:   10.50,
		DirectCost:  5.25,
		Endpoints:   make(map[string]*EndpointCost),
	}

	report.AddServiceCost(serviceCost)

	if len(report.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(report.Services))
	}

	if report.TotalCost != 10.50 {
		t.Errorf("Expected total cost 10.50, got %f", report.TotalCost)
	}
}

func TestCalculateTotalCost(t *testing.T) {
	report := NewCostReport(&CostModel{Provider: "aws"}, TimeRange{})

	services := []*ServiceCost{
		{ServiceName: "service-1", TotalCost: 10.0, Endpoints: make(map[string]*EndpointCost)},
		{ServiceName: "service-2", TotalCost: 20.0, Endpoints: make(map[string]*EndpointCost)},
		{ServiceName: "service-3", TotalCost: 15.5, Endpoints: make(map[string]*EndpointCost)},
	}

	for _, sc := range services {
		report.AddServiceCost(sc)
	}

	report.CalculateTotalCost()

	expectedTotal := 45.5
	if report.TotalCost != expectedTotal {
		t.Errorf("Expected total cost %f, got %f", expectedTotal, report.TotalCost)
	}
}

func TestNewCostBreakdown(t *testing.T) {
	metrics := &ResourceMetrics{
		CPUCores:     2.0,
		MemoryMB:     2048.0,
		NetworkInMB:  100.0,
		NetworkOutMB: 50.0,
	}

	perfMetrics := &PerformanceMetrics{
		RequestRate: 100.0,
	}

	model := &CostModel{
		CPUCostPerCoreHour:  0.05,
		MemoryCostPerGBHour: 0.01,
		NetworkCostPerGB:    0.09,
		RequestCost:         0.0000002,
	}

	durationHours := 1.0

	breakdown := NewCostBreakdown(metrics, perfMetrics, model, durationHours)

	if breakdown == nil {
		t.Fatal("NewCostBreakdown returned nil")
	}

	// CPU cost: 2.0 cores * 0.05 * 1 hour = 0.10
	expectedCPUCost := 0.10
	if breakdown.CPUCost != expectedCPUCost {
		t.Errorf("Expected CPU cost %f, got %f", expectedCPUCost, breakdown.CPUCost)
	}

	// Memory cost: (2048 MB / 1024) GB * 0.01 * 1 hour = 0.02
	expectedMemoryCost := 0.02
	if breakdown.MemoryCost != expectedMemoryCost {
		t.Errorf("Expected memory cost %f, got %f", expectedMemoryCost, breakdown.MemoryCost)
	}

	// Network cost: ((100 + 50) MB / 1024) GB * 0.09 = 0.013183...
	expectedNetworkCost := 0.013183593750000001
	tolerance := 0.0001
	if breakdown.NetworkCost < expectedNetworkCost-tolerance || breakdown.NetworkCost > expectedNetworkCost+tolerance {
		t.Errorf("Expected network cost around %f, got %f", expectedNetworkCost, breakdown.NetworkCost)
	}

	// Request cost: 100 req/s * 0.0000002 * 3600 s = 0.072
	expectedRequestCost := 0.072
	if breakdown.RequestCost != expectedRequestCost {
		t.Errorf("Expected request cost %f, got %f", expectedRequestCost, breakdown.RequestCost)
	}

	// Total should be sum of all
	if breakdown.Total <= 0 {
		t.Error("Total cost should be greater than 0")
	}
}

func TestEndpointCost(t *testing.T) {
	ec := &EndpointCost{
		Service:        "test-service",
		Endpoint:       "/api/test",
		Method:         "GET",
		DirectCost:     5.0,
		TotalCost:      10.0,
		RequestCount:   1000.0,
		CostPerRequest: 0.01,
	}

	if ec.Service != "test-service" {
		t.Errorf("Expected service 'test-service', got '%s'", ec.Service)
	}

	if ec.DirectCost != 5.0 {
		t.Errorf("Expected direct cost 5.0, got %f", ec.DirectCost)
	}

	if ec.TotalCost != 10.0 {
		t.Errorf("Expected total cost 10.0, got %f", ec.TotalCost)
	}
}

func TestDownstreamCost(t *testing.T) {
	dc := DownstreamCost{
		Service:         "downstream-service",
		Endpoint:        "/api/downstream",
		Cost:            2.5,
		CallsPerRequest: 1.5,
		Depth:           2,
	}

	if dc.Service != "downstream-service" {
		t.Errorf("Expected service 'downstream-service', got '%s'", dc.Service)
	}

	if dc.Cost != 2.5 {
		t.Errorf("Expected cost 2.5, got %f", dc.Cost)
	}

	if dc.Depth != 2 {
		t.Errorf("Expected depth 2, got %d", dc.Depth)
	}
}
