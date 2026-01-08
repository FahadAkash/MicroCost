package models

import "time"

// CostModel represents pricing for different resource types
type CostModel struct {
	CPUCostPerCoreHour  float64 `json:"cpu_cost_per_core_hour" yaml:"cpu_cost_per_core_hour"`
	MemoryCostPerGBHour float64 `json:"memory_cost_per_gb_hour" yaml:"memory_cost_per_gb_hour"`
	NetworkCostPerGB    float64 `json:"network_cost_per_gb" yaml:"network_cost_per_gb"`
	DiskCostPerGBHour   float64 `json:"disk_cost_per_gb_hour" yaml:"disk_cost_per_gb_hour"`
	RequestCost         float64 `json:"request_cost" yaml:"request_cost"`
	Provider            string  `json:"provider" yaml:"provider"`
	Region              string  `json:"region" yaml:"region"`
}

// EndpointCost represents the cost breakdown for a single endpoint
type EndpointCost struct {
	Service         string           `json:"service" yaml:"service"`
	Endpoint        string           `json:"endpoint" yaml:"endpoint"`
	Method          string           `json:"method" yaml:"method"`
	DirectCost      float64          `json:"direct_cost" yaml:"direct_cost"`
	DownstreamCosts []DownstreamCost `json:"downstream_costs" yaml:"downstream_costs"`
	TotalCost       float64          `json:"total_cost" yaml:"total_cost"`
	CostPerRequest  float64          `json:"cost_per_request" yaml:"cost_per_request"`
	RequestCount    float64          `json:"request_count" yaml:"request_count"`
	CostBreakdown   *CostBreakdown   `json:"cost_breakdown" yaml:"cost_breakdown"`
}

// DownstreamCost represents cost attributed from a downstream service
type DownstreamCost struct {
	Service         string  `json:"service" yaml:"service"`
	Endpoint        string  `json:"endpoint" yaml:"endpoint"`
	Cost            float64 `json:"cost" yaml:"cost"`
	CallsPerRequest float64 `json:"calls_per_request" yaml:"calls_per_request"`
	Depth           int     `json:"depth" yaml:"depth"` // depth in call chain
}

// CostBreakdown represents detailed cost attribution
type CostBreakdown struct {
	CPUCost         float64            `json:"cpu_cost" yaml:"cpu_cost"`
	MemoryCost      float64            `json:"memory_cost" yaml:"memory_cost"`
	NetworkCost     float64            `json:"network_cost" yaml:"network_cost"`
	DiskCost        float64            `json:"disk_cost" yaml:"disk_cost"`
	RequestCost     float64            `json:"request_cost" yaml:"request_cost"`
	DownstreamTotal float64            `json:"downstream_total" yaml:"downstream_total"`
	Total           float64            `json:"total" yaml:"total"`
	Details         map[string]float64 `json:"details,omitempty" yaml:"details,omitempty"`
}

// ServiceCost aggregates costs for all endpoints in a service
type ServiceCost struct {
	ServiceName    string                   `json:"service_name" yaml:"service_name"`
	Endpoints      map[string]*EndpointCost `json:"endpoints" yaml:"endpoints"`
	TotalCost      float64                  `json:"total_cost" yaml:"total_cost"`
	DirectCost     float64                  `json:"direct_cost" yaml:"direct_cost"`
	AttributedCost float64                  `json:"attributed_cost" yaml:"attributed_cost"`
}

// CostReport represents the complete cost analysis
type CostReport struct {
	Services        map[string]*ServiceCost `json:"services" yaml:"services"`
	TotalCost       float64                 `json:"total_cost" yaml:"total_cost"`
	GeneratedAt     time.Time               `json:"generated_at" yaml:"generated_at"`
	TimeRange       TimeRange               `json:"time_range" yaml:"time_range"`
	CostModel       *CostModel              `json:"cost_model" yaml:"cost_model"`
	TopCostly       []*EndpointCost         `json:"top_costly,omitempty" yaml:"top_costly,omitempty"`
	Recommendations []string                `json:"recommendations,omitempty" yaml:"recommendations,omitempty"`
}

// NewCostReport creates a new cost report
func NewCostReport(costModel *CostModel, timeRange TimeRange) *CostReport {
	return &CostReport{
		Services:        make(map[string]*ServiceCost),
		GeneratedAt:     time.Now(),
		TimeRange:       timeRange,
		CostModel:       costModel,
		TopCostly:       make([]*EndpointCost, 0),
		Recommendations: make([]string, 0),
	}
}

// AddServiceCost adds a service cost to the report
func (cr *CostReport) AddServiceCost(sc *ServiceCost) {
	if cr.Services == nil {
		cr.Services = make(map[string]*ServiceCost)
	}
	cr.Services[sc.ServiceName] = sc
	cr.TotalCost += sc.TotalCost
}

// CalculateTotalCost recalculates the total cost
func (cr *CostReport) CalculateTotalCost() {
	total := 0.0
	for _, sc := range cr.Services {
		total += sc.TotalCost
	}
	cr.TotalCost = total
}

// NewCostBreakdown creates a new cost breakdown from metrics
func NewCostBreakdown(metrics *ResourceMetrics, perfMetrics *PerformanceMetrics, model *CostModel, durationHours float64) *CostBreakdown {
	cb := &CostBreakdown{
		Details: make(map[string]float64),
	}

	if metrics != nil && model != nil {
		cb.CPUCost = metrics.CPUCores * model.CPUCostPerCoreHour * durationHours
		cb.MemoryCost = (metrics.MemoryMB / 1024.0) * model.MemoryCostPerGBHour * durationHours
		cb.NetworkCost = ((metrics.NetworkInMB + metrics.NetworkOutMB) / 1024.0) * model.NetworkCostPerGB
		cb.DiskCost = ((metrics.DiskReadMB + metrics.DiskWriteMB) / 1024.0) * model.DiskCostPerGBHour * durationHours
	}

	if perfMetrics != nil && model != nil {
		cb.RequestCost = perfMetrics.RequestRate * model.RequestCost * durationHours * 3600
	}

	cb.Total = cb.CPUCost + cb.MemoryCost + cb.NetworkCost + cb.DiskCost + cb.RequestCost
	return cb
}
