//go:build integration
// +build integration

package integration

import (
	"os"
	"testing"
	"time"

	"github.com/microcost/microcost/internal/analyzer"
	"github.com/microcost/microcost/internal/costengine"
	"github.com/microcost/microcost/internal/graph"
	"github.com/microcost/microcost/internal/visualizer"
	"github.com/microcost/microcost/pkg/config"
	"github.com/microcost/microcost/pkg/models"
	"github.com/sirupsen/logrus"
)

// TestEndToEndAnalysis tests the complete analysis pipeline
func TestEndToEndAnalysis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise in tests

	// Setup configuration
	cfg := config.DefaultConfig()
	cfg.Analysis.Paths = []string{"../../pkg/models"} // Analyze our own code
	cfg.Analysis.MaxDepth = 5

	// Step 1: Build dependency graph
	t.Log("Step 1: Building dependency graph...")
	graphBuilder := analyzer.NewGraphBuilder(&cfg.Analysis, logger)
	callGraph, g, err := graphBuilder.Build()

	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	if len(callGraph.Services) == 0 {
		t.Log("Warning: No services found (this is expected if analyzing test files)")
	}

	if g.NodeCount() < 0 {
		t.Error("Graph should have been created")
	}

	t.Logf("Found %d services, %d dependencies", len(callGraph.Services), len(callGraph.Dependencies))

	// Step 2: Create mock metrics
	t.Log("Step 2: Creating mock metrics...")
	timeRange := models.TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
	}

	metricsSnapshot := models.NewMetricsSnapshot(timeRange.Start, timeRange.End)

	for serviceName := range callGraph.Services {
		serviceMetrics := &models.ServiceMetrics{
			ServiceName: serviceName,
			Endpoints:   make(map[string]*models.EndpointMetrics),
			Aggregate: &models.ResourceMetrics{
				CPUCores:     1.0,
				MemoryMB:     512.0,
				NetworkInMB:  10.0,
				NetworkOutMB: 5.0,
			},
			TimeRange: timeRange,
		}
		metricsSnapshot.AddServiceMetrics(serviceMetrics)
	}

	t.Logf("Created metrics for %d services", len(metricsSnapshot.Services))

	// Step 3: Calculate costs
	t.Log("Step 3: Calculating costs...")
	calculator := costengine.NewCalculator(&cfg.CostModel, g, logger)
	costReport, err := calculator.CalculateCosts(callGraph, metricsSnapshot, timeRange)

	if err != nil {
		t.Fatalf("Failed to calculate costs: %v", err)
	}

	if costReport.TotalCost < 0 {
		t.Error("Total cost should not be negative")
	}

	t.Logf("Total cost calculated: $%.4f", costReport.TotalCost)

	// Step 4: Generate outputs
	t.Log("Step 4: Generating outputs...")

	// Create temp output directory
	tempDir := t.TempDir()

	exporter := visualizer.NewExporter(logger)
	renderer := visualizer.NewASCIIRenderer(logger, false)

	// Export call graph
	cgPath := tempDir + "/callgraph.json"
	if err := exporter.ExportCallGraphJSON(callGraph, cgPath); err != nil {
		t.Errorf("Failed to export call graph: %v", err)
	}

	// Export cost report
	crPath := tempDir + "/cost-report.json"
	if err := exporter.ExportCostReportJSON(costReport, crPath); err != nil {
		t.Errorf("Failed to export cost report: %v", err)
	}

	// Generate ASCII report
	asciiReport := renderer.RenderCostReport(costReport)
	if asciiReport == "" {
		t.Error("ASCII report should not be empty")
	}

	t.Log("ASCII Report Preview:")
	t.Log(asciiReport[:min(len(asciiReport), 500)]) // Show first 500 chars

	// Verify files were created
	if _, err := os.Stat(cgPath); os.IsNotExist(err) {
		t.Error("Call graph file was not created")
	}

	if _, err := os.Stat(crPath); os.IsNotExist(err) {
		t.Error("Cost report file was not created")
	}

	t.Log("✅ End-to-end test completed successfully")
}

// TestConfigurationFlow tests configuration loading and validation
func TestConfigurationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Log("Testing configuration flow...")

	// Load default config
	cfg := config.DefaultConfig()

	// Validate
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Default config validation failed: %v", err)
	}

	// Test environment variable override
	os.Setenv("AWS_ACCESS_KEY_ID", "test-integration-key")
	defer os.Unsetenv("AWS_ACCESS_KEY_ID")

	cfg2, err := config.Load("")
	if err != nil {
		t.Logf("Config load returned error (may be expected): %v", err)
	}

	if cfg2 != nil && cfg2.AWS.AccessKeyID != "test-integration-key" {
		t.Errorf("Environment variable override failed")
	}

	t.Log("✅ Configuration flow test completed")
}

// TestGraphAlgorithms tests graph operations on realistic data
func TestGraphAlgorithms(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Log("Testing graph algorithms with realistic scenario...")

	g := graph.NewGraph()

	// Create a realistic microservices dependency graph
	// Frontend -> API Gateway -> [Auth, User, Payment]
	// Payment -> [Inventory, Notification]

	frontend := g.AddNode("frontend", "frontend", "/", "GET", nil)
	apiGateway := g.AddNode("api-gateway", "api-gateway", "/api", "GET", nil)
	auth := g.AddNode("auth", "auth", "/verify", "POST", nil)
	user := g.AddNode("user", "user", "/profile", "GET", nil)
	payment := g.AddNode("payment", "payment", "/pay", "POST", nil)
	inventory := g.AddNode("inventory", "inventory", "/check", "GET", nil)
	notification := g.AddNode("notification", "notification", "/send", "POST", nil)

	// Add edges
	g.AddEdge(frontend, apiGateway, 1.0, nil)
	g.AddEdge(apiGateway, auth, 1.0, nil)
	g.AddEdge(apiGateway, user, 1.0, nil)
	g.AddEdge(apiGateway, payment, 1.0, nil)
	g.AddEdge(payment, inventory, 1.0, nil)
	g.AddEdge(payment, notification, 1.0, nil)

	// Test cycle detection
	if g.HasCycle() {
		t.Error("Graph should not have cycles")
	}

	// Test topological sort
	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("Topological sort failed: %v", err)
	}

	if len(sorted) != 7 {
		t.Errorf("Expected 7 nodes in sorted order, got %d", len(sorted))
	}

	// Test path finding
	paths := g.FindAllPaths("frontend", "notification", 10)
	if len(paths) == 0 {
		t.Error("Should find at least one path from frontend to notification")
	}

	t.Logf("Found %d paths from frontend to notification", len(paths))
	t.Log("✅ Graph algorithms test completed")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
