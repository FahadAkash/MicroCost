package models

import "time"

// ResourceMetrics represents resource consumption data
type ResourceMetrics struct {
	CPUCores     float64   `json:"cpu_cores" yaml:"cpu_cores"`
	MemoryMB     float64   `json:"memory_mb" yaml:"memory_mb"`
	NetworkInMB  float64   `json:"network_in_mb" yaml:"network_in_mb"`
	NetworkOutMB float64   `json:"network_out_mb" yaml:"network_out_mb"`
	DiskReadMB   float64   `json:"disk_read_mb" yaml:"disk_read_mb"`
	DiskWriteMB  float64   `json:"disk_write_mb" yaml:"disk_write_mb"`
	Timestamp    time.Time `json:"timestamp" yaml:"timestamp"`
}

// PerformanceMetrics represents performance-related metrics
type PerformanceMetrics struct {
	RequestRate float64       `json:"request_rate" yaml:"request_rate"` // req/sec
	ErrorRate   float64       `json:"error_rate" yaml:"error_rate"`     // errors/sec
	LatencyAvg  time.Duration `json:"latency_avg" yaml:"latency_avg"`
	LatencyP50  time.Duration `json:"latency_p50" yaml:"latency_p50"`
	LatencyP95  time.Duration `json:"latency_p95" yaml:"latency_p95"`
	LatencyP99  time.Duration `json:"latency_p99" yaml:"latency_p99"`
	Timestamp   time.Time     `json:"timestamp" yaml:"timestamp"`
}

// EndpointMetrics represents combined metrics for an endpoint
type EndpointMetrics struct {
	Service     string              `json:"service" yaml:"service"`
	Endpoint    string              `json:"endpoint" yaml:"endpoint"`
	Method      string              `json:"method" yaml:"method"`
	Resource    *ResourceMetrics    `json:"resource" yaml:"resource"`
	Performance *PerformanceMetrics `json:"performance" yaml:"performance"`
	TimeRange   TimeRange           `json:"time_range" yaml:"time_range"`
}

// ServiceMetrics aggregates metrics for all endpoints in a service
type ServiceMetrics struct {
	ServiceName string                      `json:"service_name" yaml:"service_name"`
	Endpoints   map[string]*EndpointMetrics `json:"endpoints" yaml:"endpoints"`
	Aggregate   *ResourceMetrics            `json:"aggregate" yaml:"aggregate"`
	TimeRange   TimeRange                   `json:"time_range" yaml:"time_range"`
}

// TimeRange represents a time window for metrics
type TimeRange struct {
	Start time.Time `json:"start" yaml:"start"`
	End   time.Time `json:"end" yaml:"end"`
}

// MetricsSnapshot represents a point-in-time snapshot of all metrics
type MetricsSnapshot struct {
	Services   map[string]*ServiceMetrics `json:"services" yaml:"services"`
	CapturedAt time.Time                  `json:"captured_at" yaml:"captured_at"`
	TimeRange  TimeRange                  `json:"time_range" yaml:"time_range"`
}

// NewMetricsSnapshot creates a new metrics snapshot
func NewMetricsSnapshot(start, end time.Time) *MetricsSnapshot {
	return &MetricsSnapshot{
		Services:   make(map[string]*ServiceMetrics),
		CapturedAt: time.Now(),
		TimeRange: TimeRange{
			Start: start,
			End:   end,
		},
	}
}

// AddServiceMetrics adds service metrics to the snapshot
func (ms *MetricsSnapshot) AddServiceMetrics(sm *ServiceMetrics) {
	if ms.Services == nil {
		ms.Services = make(map[string]*ServiceMetrics)
	}
	ms.Services[sm.ServiceName] = sm
}

// GetServiceMetrics retrieves metrics for a specific service
func (ms *MetricsSnapshot) GetServiceMetrics(serviceName string) (*ServiceMetrics, bool) {
	sm, exists := ms.Services[serviceName]
	return sm, exists
}
