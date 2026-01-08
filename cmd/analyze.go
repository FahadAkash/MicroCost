package cmd

import (
	"github.com/microcost/microcost/internal/analyzer"
	"github.com/microcost/microcost/internal/visualizer"
	"github.com/microcost/microcost/pkg/config"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze codebase and build dependency graph",
	Long: `Scans your Go codebase to discover services, detect HTTP and gRPC calls,
and build a complete dependency graph of your microservices architecture.`,
	RunE: runAnalyze,
}

var (
	analyzePaths     []string
	analyzeOutput    string
	analyzeFormat    string
	analyzeVisualize bool
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringSliceVarP(&analyzePaths, "paths", "p", nil, "Paths to analyze")
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "callgraph.json", "Output file path")
	analyzeCmd.Flags().StringVarP(&analyzeFormat, "format", "f", "json", "Output format (json, yaml)")
	analyzeCmd.Flags().BoolVarP(&analyzeVisualize, "visualize", "v", true, "Show ASCII visualization")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	logger := GetLogger()
	logger.Info("Starting code analysis...")

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		logger.WithError(err).Warn("Error loading config, using defaults")
		cfg = config.DefaultConfig()
	}

	// Override paths if provided
	if len(analyzePaths) > 0 {
		cfg.Analysis.Paths = analyzePaths
	}

	// Build dependency graph
	graphBuilder := analyzer.NewGraphBuilder(&cfg.Analysis, logger)
	callGraph, _, err := graphBuilder.Build()
	if err != nil {
		logger.WithError(err).Error("Error building dependency graph")
		return err
	}

	logger.Infof("Analysis complete: %d services, %d dependencies",
		len(callGraph.Services), len(callGraph.Dependencies))

	// Export to file
	exporter := visualizer.NewExporter(logger)
	if analyzeFormat == "yaml" {
		err = exporter.ExportYAML(callGraph, analyzeOutput)
	} else {
		err = exporter.ExportCallGraphJSON(callGraph, analyzeOutput)
	}

	if err != nil {
		logger.WithError(err).Error("Error exporting call graph")
		return err
	}

	logger.Infof("Call graph exported to: %s", analyzeOutput)

	// Show ASCII visualization if requested
	if analyzeVisualize {
		renderer := visualizer.NewASCIIRenderer(logger, cfg.Output.ColorEnabled)

		// Show dependency tree for the first service
		for serviceName := range callGraph.Services {
			treeOutput := renderer.RenderDependencyTree(callGraph, serviceName)
			cmd.Println(treeOutput)
			break // Just show first service for now
		}
	}

	logger.Info("✓ Analysis complete")
	return nil
}
