package costengine

import (
	"fmt"

	"github.com/microcost/microcost/internal/graph"
	"github.com/microcost/microcost/pkg/config"
	"github.com/microcost/microcost/pkg/models"
	"github.com/sirupsen/logrus"
)

// Calculator calculates costs for services and endpoints
type Calculator struct {
	config    *config.CostModelConfig
	logger    *logrus.Logger
	costModel *models.CostModel
	graph     *graph.Graph
}

// NewCalculator creates a new cost calculator
func NewCalculator(cfg *config.CostModelConfig, g *graph.Graph, logger *logrus.Logger) *Calculator {
	costModel := &models.CostModel{
		CPUCostPerCoreHour:  cfg.CPUCostPerCoreHour,
		MemoryCostPerGBHour: cfg.MemoryCostPerGBHour,
		NetworkCostPerGB:    cfg.NetworkCostPerGB,
		DiskCostPerGBHour:   cfg.DiskCostPerGBHour,
		RequestCost:         cfg.RequestCost,
		Provider:            cfg.Provider,
		Region:              cfg.Region,
	}

	return &Calculator{
		config:    cfg,
		logger:    logger,
		costModel: costModel,
		graph:     g,
	}
}

// CalculateCosts calculates costs for all services and endpoints
func (c *Calculator) CalculateCosts(callGraph *models.CallGraph, metricsSnapshot *models.MetricsSnapshot, timeRange models.TimeRange) (*models.CostReport, error) {
	c.logger.Info("Calculating costs...")

	report := models.NewCostReport(c.costModel, timeRange)

	// Calculate duration in hours for cost calculation
	durationHours := timeRange.End.Sub(timeRange.Start).Hours()

	// Calculate costs for each service
	for serviceName, service := range callGraph.Services {
		serviceCost := &models.ServiceCost{
			ServiceName: serviceName,
			Endpoints:   make(map[string]*models.EndpointCost),
		}

		// Get service metrics
		serviceMetrics, _ := metricsSnapshot.GetServiceMetrics(serviceName)

		// Calculate costs for each endpoint
		for _, endpoint := range service.Endpoints {
			endpointCost := c.calculateEndpointCost(endpoint, serviceMetrics, durationHours)

			key := fmt.Sprintf("%s:%s", endpoint.Path, endpoint.Method)
			serviceCost.Endpoints[key] = endpointCost
			serviceCost.DirectCost += endpointCost.DirectCost
		}

		// Calculate attributed costs (downstream dependencies)
		for _, endpoint := range service.Endpoints {
			key := fmt.Sprintf("%s:%s", endpoint.Path, endpoint.Method)
			endpointCost := serviceCost.Endpoints[key]

			downstreamCosts := c.calculateDownstreamCosts(endpoint, callGraph, serviceCost.Endpoints, 0, make(map[string]bool))
			endpointCost.DownstreamCosts = downstreamCosts

			// Sum up downstream costs
			downstreamTotal := 0.0
			for _, dc := range downstreamCosts {
				downstreamTotal += dc.Cost
			}

			endpointCost.TotalCost = endpointCost.DirectCost + downstreamTotal
			if endpointCost.CostBreakdown != nil {
				endpointCost.CostBreakdown.DownstreamTotal = downstreamTotal
				endpointCost.CostBreakdown.Total = endpointCost.TotalCost
			}

			// Calculate cost per request
			if endpointCost.RequestCount > 0 {
				endpointCost.CostPerRequest = endpointCost.TotalCost / endpointCost.RequestCount
			}
		}

		// Calculate total service cost
		serviceCost.TotalCost = serviceCost.DirectCost
		for _, ec := range serviceCost.Endpoints {
			downstreamTotal := 0.0
			for _, dc := range ec.DownstreamCosts {
				downstreamTotal += dc.Cost
			}
			serviceCost.AttributedCost += downstreamTotal
		}
		serviceCost.TotalCost += serviceCost.AttributedCost

		report.AddServiceCost(serviceCost)
	}

	// Find top costly endpoints
	report.TopCostly = c.findTopCostlyEndpoints(report, 10)

	// Generate recommendations
	report.Recommendations = c.generateRecommendations(report)

	c.logger.Infof("Cost calculation complete. Total cost: $%.2f", report.TotalCost)
	return report, nil
}

// calculateEndpointCost calculates the direct cost for an endpoint
func (c *Calculator) calculateEndpointCost(endpoint *models.Endpoint, serviceMetrics *models.ServiceMetrics, durationHours float64) *models.EndpointCost {
	ec := &models.EndpointCost{
		Service:  endpoint.Service.Name,
		Endpoint: endpoint.Path,
		Method:   endpoint.Method,
	}

	if serviceMetrics == nil {
		return ec
	}

	key := fmt.Sprintf("%s:%s", endpoint.Path, endpoint.Method)
	endpointMetrics, exists := serviceMetrics.Endpoints[key]
	if !exists || endpointMetrics.Resource == nil {
		return ec
	}

	// Calculate cost breakdown
	costBreakdown := models.NewCostBreakdown(
		endpointMetrics.Resource,
		endpointMetrics.Performance,
		c.costModel,
		durationHours,
	)

	ec.DirectCost = costBreakdown.Total
	ec.CostBreakdown = costBreakdown

	// Store request count for cost per request calculation
	if endpointMetrics.Performance != nil {
		ec.RequestCount = endpointMetrics.Performance.RequestRate * durationHours * 3600
	}

	return ec
}

// calculateDownstreamCosts recursively calculates costs from downstream dependencies
func (c *Calculator) calculateDownstreamCosts(endpoint *models.Endpoint, callGraph *models.CallGraph, endpointCosts map[string]*models.EndpointCost, depth int, visited map[string]bool) []models.DownstreamCost {
	maxDepth := 10 // Prevent infinite recursion
	if depth > maxDepth {
		return nil
	}

	downstreamCosts := make([]models.DownstreamCost, 0)

	// Find dependencies for this endpoint
	for _, dep := range callGraph.Dependencies {
		if dep.FromService == endpoint.Service.Name && dep.FromEndpoint == endpoint.Path {
			// Check if we've already visited this dependency (cycle detection)
			depKey := fmt.Sprintf("%s:%s", dep.ToService, dep.ToEndpoint)
			if visited[depKey] {
				c.logger.Warnf("Circular dependency detected: %s", depKey)
				continue
			}

			// Find the cost of the downstream endpoint
			targetKey := fmt.Sprintf("%s:%s", dep.ToEndpoint, "GET") // Simplified
			var targetCost float64

			if ec, exists := endpointCosts[targetKey]; exists {
				targetCost = ec.DirectCost
			}

			// Apply weight (calls per request)
			weightedCost := targetCost * dep.Weight

			dc := models.DownstreamCost{
				Service:         dep.ToService,
				Endpoint:        dep.ToEndpoint,
				Cost:            weightedCost,
				CallsPerRequest: dep.Weight,
				Depth:           depth + 1,
			}

			downstreamCosts = append(downstreamCosts, dc)

			// Recursively calculate downstream costs of the dependency
			if targetService, exists := callGraph.GetService(dep.ToService); exists {
				if targetEndpoint, epExists := targetService.GetEndpoint(dep.ToEndpoint, "GET"); epExists {
					visited[depKey] = true
					nestedCosts := c.calculateDownstreamCosts(targetEndpoint, callGraph, endpointCosts, depth+1, visited)
					delete(visited, depKey)

					// Add nested costs (scaled by weight)
					for _, nc := range nestedCosts {
						nc.Cost *= dep.Weight
						downstreamCosts = append(downstreamCosts, nc)
					}
				}
			}
		}
	}

	return downstreamCosts
}

// findTopCostlyEndpoints finds the most expensive endpoints
func (c *Calculator) findTopCostlyEndpoints(report *models.CostReport, n int) []*models.EndpointCost {
	allEndpoints := make([]*models.EndpointCost, 0)

	for _, serviceCost := range report.Services {
		for _, endpointCost := range serviceCost.Endpoints {
			allEndpoints = append(allEndpoints, endpointCost)
		}
	}

	// Simple bubble sort for top N (in production, use a heap)
	for i := 0; i < len(allEndpoints)-1; i++ {
		for j := i + 1; j < len(allEndpoints); j++ {
			if allEndpoints[j].TotalCost > allEndpoints[i].TotalCost {
				allEndpoints[i], allEndpoints[j] = allEndpoints[j], allEndpoints[i]
			}
		}
	}

	if len(allEndpoints) > n {
		return allEndpoints[:n]
	}

	return allEndpoints
}

// generateRecommendations generates cost optimization recommendations
func (c *Calculator) generateRecommendations(report *models.CostReport) []string {
	recommendations := make([]string, 0)

	// Analyze top costly endpoints
	if len(report.TopCostly) > 0 {
		topEndpoint := report.TopCostly[0]
		recommendations = append(recommendations,
			fmt.Sprintf("Consider optimizing %s%s - highest cost endpoint at $%.4f per request",
				topEndpoint.Service, topEndpoint.Endpoint, topEndpoint.CostPerRequest))
	}

	// Check for endpoints with high downstream costs
	for _, serviceCost := range report.Services {
		for _, ec := range serviceCost.Endpoints {
			if ec.CostBreakdown != nil && ec.DirectCost > 0 {
				ratio := ec.CostBreakdown.DownstreamTotal / ec.DirectCost
				if ratio > 5.0 {
					recommendations = append(recommendations,
						fmt.Sprintf("%s%s has 5x more downstream cost than direct cost - review dependency chain",
							ec.Service, ec.Endpoint))
				}
			}
		}
	}

	return recommendations
}

// GetCostModel returns the cost model being used
func (c *Calculator) GetCostModel() *models.CostModel {
	return c.costModel
}
