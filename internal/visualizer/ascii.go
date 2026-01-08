package visualizer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/microcost/microcost/pkg/models"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
)

// ASCIIRenderer renders dependency graphs and cost reports in ASCII
type ASCIIRenderer struct {
	logger       *logrus.Logger
	colorEnabled bool
}

// NewASCIIRenderer creates a new ASCII renderer
func NewASCIIRenderer(logger *logrus.Logger, colorEnabled bool) *ASCIIRenderer {
	return &ASCIIRenderer{
		logger:       logger,
		colorEnabled: colorEnabled,
	}
}

// RenderCostReport renders a cost report as ASCII tree and tables
func (ar *ASCIIRenderer) RenderCostReport(report *models.CostReport) string {
	var sb strings.Builder

	// Header
	sb.WriteString(ar.renderHeader("MICROSERVICES COST REPORT"))
	sb.WriteString("\n\n")

	// Summary
	sb.WriteString(ar.renderSummary(report))
	sb.WriteString("\n\n")

	// Top Costly Endpoints
	sb.WriteString(ar.renderTopCostly(report))
	sb.WriteString("\n\n")

	// Service Breakdown
	sb.WriteString(ar.renderServiceBreakdown(report))
	sb.WriteString("\n\n")

	// Recommendations
	if len(report.Recommendations) > 0 {
		sb.WriteString(ar.renderRecommendations(report))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderHeader renders a styled header
func (ar *ASCIIRenderer) renderHeader(title string) string {
	if !ar.colorEnabled {
		return "=== " + title + " ==="
	}

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Background(lipgloss.Color("235")).
		Padding(0, 2)

	return style.Render(title)
}

// renderSummary renders the cost summary
func (ar *ASCIIRenderer) renderSummary(report *models.CostReport) string {
	var sb strings.Builder

	sb.WriteString(ar.styleLabel("Total Cost:") + " " + ar.styleCost(report.TotalCost) + "\n")
	sb.WriteString(ar.styleLabel("Time Range:") + " " + report.TimeRange.Start.Format("2006-01-02 15:04") +
		" to " + report.TimeRange.End.Format("2006-01-02 15:04") + "\n")
	sb.WriteString(ar.styleLabel("Services:") + fmt.Sprintf(" %d\n", len(report.Services)))
	sb.WriteString(ar.styleLabel("Provider:") + " " + report.CostModel.Provider + " (" + report.CostModel.Region + ")\n")

	return sb.String()
}

// renderTopCostly renders the top costly endpoints table
func (ar *ASCIIRenderer) renderTopCostly(report *models.CostReport) string {
	if len(report.TopCostly) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(ar.renderSubHeader("Top Costly Endpoints") + "\n\n")

	tableStr := &strings.Builder{}
	table := tablewriter.NewWriter(tableStr)
	table.SetHeader([]string{"Rank", "Service", "Endpoint", "Direct Cost", "Downstream", "Total Cost", "$/Request"})
	table.SetBorder(true)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for i, ec := range report.TopCostly {
		downstream := 0.0
		if ec.CostBreakdown != nil {
			downstream = ec.CostBreakdown.DownstreamTotal
		}

		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			ec.Service,
			ec.Endpoint,
			fmt.Sprintf("$%.4f", ec.DirectCost),
			fmt.Sprintf("$%.4f", downstream),
			fmt.Sprintf("$%.4f", ec.TotalCost),
			fmt.Sprintf("$%.6f", ec.CostPerRequest),
		})
	}

	table.Render()
	sb.WriteString(tableStr.String())

	return sb.String()
}

// renderServiceBreakdown renders per-service cost breakdown
func (ar *ASCIIRenderer) renderServiceBreakdown(report *models.CostReport) string {
	var sb strings.Builder
	sb.WriteString(ar.renderSubHeader("Service Cost Breakdown") + "\n\n")

	for serviceName, sc := range report.Services {
		sb.WriteString(ar.styleServiceName(serviceName) + "\n")
		sb.WriteString(fmt.Sprintf("  Direct Cost: %s\n", ar.styleCost(sc.DirectCost)))
		sb.WriteString(fmt.Sprintf("  Attributed Cost: %s\n", ar.styleCost(sc.AttributedCost)))
		sb.WriteString(fmt.Sprintf("  Total Cost: %s\n", ar.styleCost(sc.TotalCost)))
		sb.WriteString(fmt.Sprintf("  Endpoints: %d\n", len(sc.Endpoints)))

		// Show top 3 endpoints for this service
		topEndpoints := ar.getTopNEndpoints(sc.Endpoints, 3)
		if len(topEndpoints) > 0 {
			sb.WriteString("  Top Endpoints:\n")
			for _, ec := range topEndpoints {
				sb.WriteString(fmt.Sprintf("    • %s (%s) - %s\n",
					ec.Endpoint, ec.Method, ar.styleCost(ec.TotalCost)))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderRecommendations renders optimization recommendations
func (ar *ASCIIRenderer) renderRecommendations(report *models.CostReport) string {
	var sb strings.Builder
	sb.WriteString(ar.renderSubHeader("💡 Recommendations") + "\n\n")

	for i, rec := range report.Recommendations {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
	}

	return sb.String()
}

// renderSubHeader renders a sub-header
func (ar *ASCIIRenderer) renderSubHeader(title string) string {
	if !ar.colorEnabled {
		return "--- " + title + " ---"
	}

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("214"))

	return style.Render("▶ " + title)
}

// styleLabel styles a label
func (ar *ASCIIRenderer) styleLabel(label string) string {
	if !ar.colorEnabled {
		return label
	}

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39"))

	return style.Render(label)
}

// styleCost styles a cost value
func (ar *ASCIIRenderer) styleCost(cost float64) string {
	costStr := fmt.Sprintf("$%.4f", cost)

	if !ar.colorEnabled {
		return costStr
	}

	var color string
	if cost > 10.0 {
		color = "196" // Red
	} else if cost > 1.0 {
		color = "214" // Orange
	} else {
		color = "46" // Green
	}

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(color))

	return style.Render(costStr)
}

// styleServiceName styles a service name
func (ar *ASCIIRenderer) styleServiceName(name string) string {
	if !ar.colorEnabled {
		return "▶ " + name
	}

	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("207"))

	return style.Render("▶ " + name)
}

// getTopNEndpoints gets the top N endpoints by cost
func (ar *ASCIIRenderer) getTopNEndpoints(endpoints map[string]*models.EndpointCost, n int) []*models.EndpointCost {
	list := make([]*models.EndpointCost, 0, len(endpoints))
	for _, ec := range endpoints {
		list = append(list, ec)
	}

	// Simple bubble sort
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].TotalCost > list[i].TotalCost {
				list[i], list[j] = list[j], list[i]
			}
		}
	}

	if len(list) > n {
		return list[:n]
	}
	return list
}

// RenderDependencyTree renders a dependency tree in ASCII
func (ar *ASCIIRenderer) RenderDependencyTree(callGraph *models.CallGraph, rootService string) string {
	var sb strings.Builder

	sb.WriteString(ar.renderHeader("DEPENDENCY TREE: " + rootService))
	sb.WriteString("\n\n")

	visited := make(map[string]bool)
	ar.renderTreeNode(&sb, callGraph, rootService, "", visited, 0, 5)

	return sb.String()
}

// renderTreeNode recursively renders a tree node
func (ar *ASCIIRenderer) renderTreeNode(sb *strings.Builder, cg *models.CallGraph, serviceName, prefix string, visited map[string]bool, depth, maxDepth int) {
	if depth > maxDepth || visited[serviceName] {
		if visited[serviceName] {
			sb.WriteString(prefix + "  (circular reference)\n")
		}
		return
	}

	visited[serviceName] = true
	defer func() { visited[serviceName] = false }()

	sb.WriteString(prefix + ar.styleServiceName(serviceName) + "\n")

	// Find dependencies
	deps := make([]*models.Dependency, 0)
	for _, dep := range cg.Dependencies {
		if dep.FromService == serviceName {
			deps = append(deps, dep)
		}
	}

	for i, dep := range deps {
		isLast := i == len(deps)-1
		var newPrefix string
		if isLast {
			sb.WriteString(prefix + "  └─ ")
			newPrefix = prefix + "     "
		} else {
			sb.WriteString(prefix + "  ├─ ")
			newPrefix = prefix + "  │  "
		}

		sb.WriteString(fmt.Sprintf("%s (%s, weight: %.1f)\n", dep.ToEndpoint, dep.CallType, dep.Weight))
		ar.renderTreeNode(sb, cg, dep.ToService, newPrefix, visited, depth+1, maxDepth)
	}
}
