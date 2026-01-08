package cmd

import (
	"encoding/json"
	"os"

	"github.com/microcost/microcost/internal/costengine"
	"github.com/microcost/microcost/internal/graph"
	"github.com/microcost/microcost/internal/visualizer"
	"github.com/microcost/microcost/pkg/config"
	"github.com/microcost/microcost/pkg/models"
	"github.com/spf13/cobra"
)

var calculateCmd = &cobra.Command{
	Use:   "calculate",
	Short: "Calculate costs for all endpoints",
	Long: `Combines dependency graph with metrics data to calculate direct and attributed costs
for every API endpoint, including downstream service costs.`,
	RunE: runCalculate,
}

var (
	calculateCallGraph string
	calculateMetrics   string
	calculateOutput    string
	calculateFormat    string
	calculateVisualize bool
)

func init() {
	rootCmd.AddCommand(calculateCmd)

	calculateCmd.Flags().StringVarP(&calculateCallGraph, "callgraph", "g", "callgraph.json", "Call graph input file")
	calculateCmd.Flags().StringVarP(&calculateMetrics, "metrics", "m", "metrics.json", "Metrics input file")
	calculateCmd.Flags().StringVarP(&calculateOutput, "output", "o", "cost-report.json", "Output file path")
	calculateCmd.Flags().StringVarP(&calculateFormat, "format", "f", "json", "Output format (json, yaml, ascii)")
	calculateCmd.Flags().BoolVarP(&calculateVisualize, "visualize", "v", true, "Show ASCII visualization")
}

func runCalculate(cmd *cobra.Command, args []string) error {
	logger := GetLogger()
	logger.Info("Starting cost calculation...")

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		logger.WithError(err).Warn("Error loading config, using defaults")
		cfg = config.DefaultConfig()
	}

	// Load call graph
	callGraph, err := loadCallGraph(calculateCallGraph)
	if err != nil {
		logger.WithError(err).Error("Error loading call graph")
		return err
	}

	// Load metrics
	metricsSnapshot, err := loadMetrics(calculateMetrics)
	if err != nil {
		logger.WithError(err).Error("Error loading metrics")
		return err
	}

	// Create graph structure
	g := graph.NewGraph()

	// Create cost calculator
	calculator := costengine.NewCalculator(&cfg.CostModel, g, logger)

	// Calculate costs
	costReport, err := calculator.CalculateCosts(callGraph, metricsSnapshot, metricsSnapshot.TimeRange)
	if err != nil {
		logger.WithError(err).Error("Error calculating costs")
		return err
	}

	logger.Infof("Cost calculation complete. Total cost: $%.2f", costReport.TotalCost)

	// Export cost report
	exporter := visualizer.NewExporter(logger)
	if calculateFormat == "yaml" {
		err = exporter.ExportYAML(costReport, calculateOutput)
	} else if calculateFormat == "ascii" || calculateVisualize {
		// Show ASCII report
		renderer := visualizer.NewASCIIRenderer(logger, cfg.Output.ColorEnabled)
		asciiReport := renderer.RenderCostReport(costReport)
		cmd.Println(asciiReport)

		if calculateFormat == "ascii" {
			return nil
		}
		err = exporter.ExportCostReportJSON(costReport, calculateOutput)
	} else {
		err = exporter.ExportCostReportJSON(costReport, calculateOutput)
	}

	if err != nil {
		logger.WithError(err).Error("Error exporting cost report")
		return err
	}

	logger.Infof("Cost report exported to: %s", calculateOutput)
	logger.Info("✓ Calculation complete")
	return nil
}

// loadCallGraph loads a call graph from a file
func loadCallGraph(path string) (*models.CallGraph, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cg models.CallGraph
	if err := json.NewDecoder(file).Decode(&cg); err != nil {
		return nil, err
	}

	return &cg, nil
}

// loadMetrics loads metrics from a file
func loadMetrics(path string) (*models.MetricsSnapshot, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ms models.MetricsSnapshot
	if err := json.NewDecoder(file).Decode(&ms); err != nil {
		return nil, err
	}

	return &ms, nil
}
