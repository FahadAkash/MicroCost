package cmd

import (
	"time"

	"github.com/microcost/microcost/internal/analyzer"
	"github.com/microcost/microcost/internal/collector"
	"github.com/microcost/microcost/internal/costengine"
	"github.com/microcost/microcost/internal/visualizer"
	"github.com/microcost/microcost/pkg/config"
	"github.com/microcost/microcost/pkg/models"
	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Run complete pipeline: analyze, collect, calculate",
	Long: `Executes the full workflow: 
1. Analyzes code to build dependency graph
2. Collects metrics from Prometheus
3. Calculates costs with attribution
4. Generates comprehensive report`,
	RunE: runAll,
}

var (
	allDuration string
	allOutput   string
)

func init() {
	rootCmd.AddCommand(allCmd)

	allCmd.Flags().StringVarP(&allDuration, "duration", "d", "1h", "Time window for metrics")
	allCmd.Flags().StringVarP(&allOutput, "output", "o", "./output", "Output directory")
}

func runAll(cmd *cobra.Command, args []string) error {
	logger := GetLogger()
	logger.Info("🚀 Starting full pipeline...")

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		logger.WithError(err).Warn("Error loading config, using defaults")
		cfg = config.DefaultConfig()
	}

	// Override output path
	if allOutput != "" {
		cfg.Output.OutputPath = allOutput
	}

	// Step 1: Analyze code
	logger.Info("Step 1/3: Analyzing codebase...")
	graphBuilder := analyzer.NewGraphBuilder(&cfg.Analysis, logger)
	callGraph, g, err := graphBuilder.Build()
	if err != nil {
		logger.WithError(err).Error("Error building dependency graph")
		return err
	}
	logger.Infof("✓ Found %d services, %d dependencies",
		len(callGraph.Services), len(callGraph.Dependencies))

	// Step 2: Collect metrics
	logger.Info("Step 2/3: Collecting metrics from Prometheus...")

	duration, err := time.ParseDuration(allDuration)
	if err != nil {
		logger.WithError(err).Error("Invalid duration")
		return err
	}

	endTime := time.Now()
	startTime := endTime.Add(-duration)
	timeRange := models.TimeRange{
		Start: startTime,
		End:   endTime,
	}

	promCollector, err := collector.NewPrometheusCollector(&cfg.Prometheus, logger)
	if err != nil {
		logger.WithError(err).Error("Error creating Prometheus collector")
		return err
	}

	metricsSnapshot, err := promCollector.CollectMetrics(callGraph.Services, timeRange)
	if err != nil {
		logger.WithError(err).Error("Error collecting metrics")
		return err
	}
	logger.Infof("✓ Collected metrics for %d services", len(metricsSnapshot.Services))

	// Step 3: Calculate costs
	logger.Info("Step 3/3: Calculating costs...")
	calculator := costengine.NewCalculator(&cfg.CostModel, g, logger)
	costReport, err := calculator.CalculateCosts(callGraph, metricsSnapshot, timeRange)
	if err != nil {
		logger.WithError(err).Error("Error calculating costs")
		return err
	}
	logger.Infof("✓ Total cost: $%.2f", costReport.TotalCost)

	// Generate outputs
	logger.Info("Generating outputs...")

	exporter := visualizer.NewExporter(logger)
	renderer := visualizer.NewASCIIRenderer(logger, cfg.Output.ColorEnabled)

	// Export call graph
	if err := exporter.ExportCallGraphJSON(callGraph, allOutput+"/callgraph.json"); err != nil {
		logger.WithError(err).Error("Error exporting call graph")
	}

	// Export metrics
	if err := exporter.ExportMetricsJSON(metricsSnapshot, allOutput+"/metrics.json"); err != nil {
		logger.WithError(err).Error("Error exporting metrics")
	}

	// Export cost report
	if err := exporter.ExportCostReportJSON(costReport, allOutput+"/cost-report.json"); err != nil {
		logger.WithError(err).Error("Error exporting cost report")
	}

	// Show ASCII report
	asciiReport := renderer.RenderCostReport(costReport)
	cmd.Println("\n" + asciiReport)

	logger.Info("✅ Pipeline complete! All outputs saved to:", allOutput)
	return nil
}
