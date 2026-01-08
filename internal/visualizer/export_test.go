package visualizer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/microcost/microcost/pkg/models"
	"github.com/sirupsen/logrus"
)

func TestNewExporter(t *testing.T) {
	logger := logrus.New()
	exporter := NewExporter(logger)

	if exporter == nil {
		t.Fatal("NewExporter returned nil")
	}

	if exporter.logger != logger {
		t.Error("Logger not set correctly")
	}
}

func TestExportJSON(t *testing.T) {
	logger := logrus.New()
	exporter := NewExporter(logger)

	// Create temp file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.json")

	testData := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}

	err := exporter.ExportJSON(testData, outputPath)

	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

func TestExportYAML(t *testing.T) {
	logger := logrus.New()
	exporter := NewExporter(logger)

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.yaml")

	testData := map[string]interface{}{
		"name":  "test",
		"value": 456,
	}

	err := exporter.ExportYAML(testData, outputPath)

	if err != nil {
		t.Fatalf("ExportYAML failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

func TestExportCallGraphJSON(t *testing.T) {
	logger := logrus.New()
	exporter := NewExporter(logger)

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "callgraph.json")

	cg := models.NewCallGraph()
	service := &models.Service{
		Name:         "test-service",
		Endpoints:    make([]*models.Endpoint, 0),
		Dependencies: make([]*models.Dependency, 0),
	}
	cg.AddService(service)

	err := exporter.ExportCallGraphJSON(cg, outputPath)

	if err != nil {
		t.Fatalf("ExportCallGraphJSON failed: %v", err)
	}

	// Verify file exists and has content
	info, err := os.Stat(outputPath)
	if os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	if info.Size() == 0 {
		t.Error("Output file is empty")
	}
}

func TestExportCostReportJSON(t *testing.T) {
	logger := logrus.New()
	exporter := NewExporter(logger)

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "cost-report.json")

	costModel := &models.CostModel{
		Provider: "aws",
	}

	report := models.NewCostReport(costModel, models.TimeRange{})

	err := exporter.ExportCostReportJSON(report, outputPath)

	if err != nil {
		t.Fatalf("ExportCostReportJSON failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

func TestExportToInvalidPath(t *testing.T) {
	logger := logrus.New()
	exporter := NewExporter(logger)

	// Try to export to an invalid path
	invalidPath := "/nonexistent/directory/file.json"

	err := exporter.ExportJSON(map[string]string{"test": "data"}, invalidPath)

	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}
