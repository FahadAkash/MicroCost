package visualizer

import (
	"testing"
	"time"

	"github.com/microcost/microcost/pkg/models"
	"github.com/sirupsen/logrus"
)

func TestNewASCIIRenderer(t *testing.T) {
	logger := logrus.New()
	renderer := NewASCIIRenderer(logger, true)

	if renderer == nil {
		t.Fatal("NewASCIIRenderer returned nil")
	}

	if renderer.logger != logger {
		t.Error("Logger not set correctly")
	}

	if !renderer.colorEnabled {
		t.Error("Color should be enabled")
	}
}

func TestRenderCostReport(t *testing.T) {
	logger := logrus.New()
	renderer := NewASCIIRenderer(logger, false) // Disable color for testing

	costModel := &models.CostModel{
		Provider: "aws",
		Region:   "us-east-1",
	}

	report := models.NewCostReport(costModel, models.TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
	})

	serviceCost := &models.ServiceCost{
		ServiceName: "test-service",
		TotalCost:   10.0,
		Endpoints:   make(map[string]*models.EndpointCost),
	}

	report.AddServiceCost(serviceCost)

	output := renderer.RenderCostReport(report)

	if output == "" {
		t.Error("Rendered output should not be empty")
	}

	if !contains(output, "MICROSERVICES COST REPORT") {
		t.Error("Output should contain report title")
	}

	if !contains(output, "test-service") {
		t.Error("Output should contain service name")
	}
}

func TestRenderDependencyTree(t *testing.T) {
	logger := logrus.New()
	renderer := NewASCIIRenderer(logger, false)

	callGraph := models.NewCallGraph()

	service := &models.Service{
		Name:         "root-service",
		Endpoints:    make([]*models.Endpoint, 0),
		Dependencies: make([]*models.Dependency, 0),
	}

	callGraph.AddService(service)

	output := renderer.RenderDependencyTree(callGraph, "root-service")

	if output == "" {
		t.Error("Rendered output should not be empty")
	}

	if !contains(output, "root-service") {
		t.Error("Output should contain root service name")
	}
}

func TestStyleCost(t *testing.T) {
	logger := logrus.New()
	renderer := NewASCIIRenderer(logger, false)

	tests := []struct {
		name string
		cost float64
	}{
		{"low cost", 0.50},
		{"medium cost", 5.00},
		{"high cost", 15.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.styleCost(tt.cost)
			if result == "" {
				t.Error("Styled cost should not be empty")
			}
		})
	}
}

func TestGetTopNEndpoints(t *testing.T) {
	logger := logrus.New()
	renderer := NewASCIIRenderer(logger, false)

	endpoints := map[string]*models.EndpointCost{
		"ep1": {Endpoint: "/api/1", TotalCost: 10.0},
		"ep2": {Endpoint: "/api/2", TotalCost: 5.0},
		"ep3": {Endpoint: "/api/3", TotalCost: 15.0},
		"ep4": {Endpoint: "/api/4", TotalCost: 8.0},
	}

	top2 := renderer.getTopNEndpoints(endpoints, 2)

	if len(top2) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(top2))
	}

	// First should be highest cost
	if top2[0].TotalCost != 15.0 {
		t.Errorf("Expected highest cost 15.0, got %f", top2[0].TotalCost)
	}

	// Second should be second highest
	if top2[1].TotalCost != 10.0 {
		t.Errorf("Expected second cost 10.0, got %f", top2[1].TotalCost)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
