package analyzer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/microcost/microcost/internal/graph"
	"github.com/microcost/microcost/pkg/config"
	"github.com/microcost/microcost/pkg/models"
	"github.com/sirupsen/logrus"
)

// GraphBuilder builds a dependency graph from analyzed code
type GraphBuilder struct {
	config       *config.AnalysisConfig
	logger       *logrus.Logger
	scanner      *Scanner
	httpDetector *HTTPDetector
	grpcDetector *GRPCDetector
	callGraph    *models.CallGraph
	graph        *graph.Graph
}

// NewGraphBuilder creates a new graph builder
func NewGraphBuilder(cfg *config.AnalysisConfig, logger *logrus.Logger) *GraphBuilder {
	return &GraphBuilder{
		config:       cfg,
		logger:       logger,
		scanner:      NewScanner(cfg, logger),
		httpDetector: NewHTTPDetector(logger),
		grpcDetector: NewGRPCDetector(logger),
		callGraph:    models.NewCallGraph(),
		graph:        graph.NewGraph(),
	}
}

// Build builds the complete dependency graph
func (gb *GraphBuilder) Build() (*models.CallGraph, *graph.Graph, error) {
	gb.logger.Info("Building dependency graph...")

	// Step 1: Scan code to discover services
	services, err := gb.scanner.Scan()
	if err != nil {
		return nil, nil, fmt.Errorf("error scanning code: %w", err)
	}

	// Add services to call graph
	for _, service := range services {
		gb.callGraph.AddService(service)
	}

	// Step 2: Detect dependencies (HTTP and gRPC calls)
	if err := gb.detectDependencies(services); err != nil {
		return nil, nil, fmt.Errorf("error detecting dependencies: %w", err)
	}

	// Step 3: Build graph structure
	gb.buildGraphStructure()

	gb.logger.Infof("Graph built: %d services, %d dependencies",
		gb.graph.NodeCount(), len(gb.callGraph.Dependencies))

	return gb.callGraph, gb.graph, nil
}

// detectDependencies detects all dependencies in the codebase
func (gb *GraphBuilder) detectDependencies(services map[string]*models.Service) error {
	for serviceName, service := range services {
		gb.logger.Debugf("Detecting dependencies for service: %s", serviceName)

		// Walk through service directory
		err := filepath.Walk(service.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories and non-Go files
			if info.IsDir() || filepath.Ext(path) != ".go" {
				return nil
			}

			// Skip test files unless configured to include them
			if !gb.config.IncludeTests && filepath.Base(path) == filepath.Base(path)[:len(filepath.Base(path))-3]+"_test.go" {
				return nil
			}

			// Detect HTTP calls
			httpDeps, err := gb.httpDetector.DetectInFile(path, serviceName)
			if err != nil {
				gb.logger.WithError(err).Warnf("Error detecting HTTP calls in %s", path)
			} else {
				for _, dep := range httpDeps {
					gb.callGraph.AddDependency(dep)
				}
			}

			// Detect gRPC calls
			grpcDeps, err := gb.grpcDetector.DetectInFile(path, serviceName)
			if err != nil {
				gb.logger.WithError(err).Warnf("Error detecting gRPC calls in %s", path)
			} else {
				for _, dep := range grpcDeps {
					gb.callGraph.AddDependency(dep)
				}
			}

			return nil
		})

		if err != nil {
			gb.logger.WithError(err).Warnf("Error walking service directory: %s", service.Path)
		}
	}

	return nil
}

// buildGraphStructure builds the graph data structure from the call graph
func (gb *GraphBuilder) buildGraphStructure() {
	// Add nodes for all services and endpoints
	for _, service := range gb.callGraph.Services {
		for _, endpoint := range service.Endpoints {
			nodeID := fmt.Sprintf("%s:%s:%s", service.Name, endpoint.Path, endpoint.Method)
			gb.graph.AddNode(nodeID, service.Name, endpoint.Path, endpoint.Method, endpoint)
		}
	}

	// Add edges for all dependencies
	for _, dep := range gb.callGraph.Dependencies {
		fromID := fmt.Sprintf("%s:%s:%s", dep.FromService, dep.FromEndpoint, "GET")
		toID := fmt.Sprintf("%s:%s:%s", dep.ToService, dep.ToEndpoint, "GET")

		fromNode, fromExists := gb.graph.GetNode(fromID)
		toNode, toExists := gb.graph.GetNode(toID)

		if !fromExists {
			// Create a virtual node for the source
			fromNode = gb.graph.AddNode(fromID, dep.FromService, dep.FromEndpoint, "GET", nil)
		}

		if !toExists {
			// Create a virtual node for the target
			toNode = gb.graph.AddNode(toID, dep.ToService, dep.ToEndpoint, "GET", nil)
		}

		gb.graph.AddEdge(fromNode, toNode, dep.Weight, dep)
	}

	// Check for cycles
	if gb.graph.HasCycle() {
		gb.logger.Warn("Dependency graph contains cycles!")
	}
}

// GetCallGraph returns the built call graph
func (gb *GraphBuilder) GetCallGraph() *models.CallGraph {
	return gb.callGraph
}

// GetGraph returns the graph structure
func (gb *GraphBuilder) GetGraph() *graph.Graph {
	return gb.graph
}
