package visualizer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/microcost/microcost/pkg/models"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Exporter exports data to various formats
type Exporter struct {
	logger *logrus.Logger
}

// NewExporter creates a new exporter
func NewExporter(logger *logrus.Logger) *Exporter {
	return &Exporter{
		logger: logger,
	}
}

// ExportJSON exports data as JSON
func (e *Exporter) ExportJSON(data interface{}, outputPath string) error {
	e.logger.Infof("Exporting to JSON: %s", outputPath)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	e.logger.Info("JSON export complete")
	return nil
}

// ExportYAML exports data as YAML
func (e *Exporter) ExportYAML(data interface{}, outputPath string) error {
	e.logger.Infof("Exporting to YAML: %s", outputPath)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("error encoding YAML: %w", err)
	}

	e.logger.Info("YAML export complete")
	return nil
}

// ExportCallGraphJSON exports call graph to JSON
func (e *Exporter) ExportCallGraphJSON(cg *models.CallGraph, outputPath string) error {
	return e.ExportJSON(cg, outputPath)
}

// ExportCostReportJSON exports cost report to JSON
func (e *Exporter) ExportCostReportJSON(report *models.CostReport, outputPath string) error {
	return e.ExportJSON(report, outputPath)
}

// ExportMetricsJSON exports metrics snapshot to JSON
func (e *Exporter) ExportMetricsJSON(metrics *models.MetricsSnapshot, outputPath string) error {
	return e.ExportJSON(metrics, outputPath)
}
