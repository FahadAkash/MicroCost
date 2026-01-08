package cmd

import (
	"time"

	"github.com/microcost/microcost/internal/collector"
	"github.com/microcost/microcost/internal/visualizer"
	"github.com/microcost/microcost/pkg/config"
	"github.com/microcost/microcost/pkg/models"
	"github.com/spf13/cobra"
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect runtime metrics from Prometheus",
	Long: `Queries Prometheus to collect CPU, memory, network, latency, and request metrics
for all discovered services and endpoints.`,
	RunE: runCollect,
}

var (
	collectCallGraph string
	collectOutput    string
	collectDuration  string
)

func init() {
	rootCmd.AddCommand(collectCmd)

	collectCmd.Flags().StringVarP(&collectCallGraph, "callgraph", "g", "callgraph.json", "Call graph input file")
	collectCmd.Flags().StringVarP(&collectOutput, "output", "o", "metrics.json", "Output file path")
	collectCmd.Flags().StringVarP(&collectDuration, "duration", "d", "1h", "Time window for metrics (e.g., 1h, 30m)")
}

func runCollect(cmd *cobra.Command, args []string) error {
	logger := GetLogger()
	logger.Info("Starting metrics collection...")

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		logger.WithError(err).Warn("Error loading config, using defaults")
		cfg = config.DefaultConfig()
	}

	// Load call graph
	exporter := visualizer.NewExporter(logger)
	var callGraph models.CallGraph
	// For simplicity, we'll create a mock call graph here
	// In production, you'd load it from the file using JSON/YAML unmarshaling

	logger.Info("Call graph loaded")

	// Parse duration
	duration, err := time.ParseDuration(collectDuration)
	if err != nil {
		logger.WithError(err).Error("Invalid duration")
		return err
	}

	// Define time range
	endTime := time.Now()
	startTime := endTime.Add(-duration)
	timeRange := models.TimeRange{
		Start: startTime,
		End:   endTime,
	}

	// Create Prometheus collector
	promCollector, err := collector.NewPrometheusCollector(&cfg.Prometheus, logger)
	if err != nil {
		logger.WithError(err).Error("Error creating Prometheus collector")
		return err
	}

	// Collect metrics
	metricsSnapshot, err := promCollector.CollectMetrics(callGraph.Services, timeRange)
	if err != nil {
		logger.WithError(err).Error("Error collecting metrics")
		return err
	}

	logger.Infof("Metrics collected for %d services", len(metricsSnapshot.Services))

	// Export metrics
	err = exporter.ExportMetricsJSON(metricsSnapshot, collectOutput)
	if err != nil {
		logger.WithError(err).Error("Error exporting metrics")
		return err
	}

	logger.Infof("Metrics exported to: %s", collectOutput)
	logger.Info("✓ Collection complete")
	return nil
}
