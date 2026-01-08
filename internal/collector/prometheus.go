package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"

	"github.com/microcost/microcost/pkg/config"
	"github.com/microcost/microcost/pkg/models"
)

// PrometheusCollector collects metrics from Prometheus
type PrometheusCollector struct {
	config *config.PrometheusConfig
	logger *logrus.Logger
	client v1.API
}

// NewPrometheusCollector creates a new Prometheus collector
func NewPrometheusCollector(cfg *config.PrometheusConfig, logger *logrus.Logger) (*PrometheusCollector, error) {
	client, err := api.NewClient(api.Config{
		Address: cfg.URL,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating Prometheus client: %w", err)
	}

	return &PrometheusCollector{
		config: cfg,
		logger: logger,
		client: v1.NewAPI(client),
	}, nil
}

// CollectMetrics collects metrics for all services
func (pc *PrometheusCollector) CollectMetrics(services map[string]*models.Service, timeRange models.TimeRange) (*models.MetricsSnapshot, error) {
	pc.logger.Info("Collecting metrics from Prometheus...")

	snapshot := models.NewMetricsSnapshot(timeRange.Start, timeRange.End)

	for serviceName, service := range services {
		pc.logger.Debugf("Collecting metrics for service: %s", serviceName)

		serviceMetrics := &models.ServiceMetrics{
			ServiceName: serviceName,
			Endpoints:   make(map[string]*models.EndpointMetrics),
			TimeRange:   timeRange,
		}

		// Collect metrics for each endpoint
		for _, endpoint := range service.Endpoints {
			endpointMetrics, err := pc.collectEndpointMetrics(serviceName, endpoint, timeRange)
			if err != nil {
				pc.logger.WithError(err).Warnf("Error collecting metrics for %s%s", serviceName, endpoint.Path)
				continue
			}

			key := fmt.Sprintf("%s:%s", endpoint.Path, endpoint.Method)
			serviceMetrics.Endpoints[key] = endpointMetrics
		}

		// Calculate aggregate service metrics
		serviceMetrics.Aggregate = pc.aggregateServiceMetrics(serviceMetrics.Endpoints)

		snapshot.AddServiceMetrics(serviceMetrics)
	}

	pc.logger.Info("Metrics collection complete")
	return snapshot, nil
}

// collectEndpointMetrics collects metrics for a specific endpoint
func (pc *PrometheusCollector) collectEndpointMetrics(service string, endpoint *models.Endpoint, timeRange models.TimeRange) (*models.EndpointMetrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pc.config.Timeout)
	defer cancel()

	metrics := &models.EndpointMetrics{
		Service:   service,
		Endpoint:  endpoint.Path,
		Method:    endpoint.Method,
		TimeRange: timeRange,
	}

	// Collect resource metrics
	resourceMetrics, err := pc.collectResourceMetrics(ctx, service, endpoint, timeRange)
	if err != nil {
		return nil, fmt.Errorf("error collecting resource metrics: %w", err)
	}
	metrics.Resource = resourceMetrics

	// Collect performance metrics
	perfMetrics, err := pc.collectPerformanceMetrics(ctx, service, endpoint, timeRange)
	if err != nil {
		return nil, fmt.Errorf("error collecting performance metrics: %w", err)
	}
	metrics.Performance = perfMetrics

	return metrics, nil
}

// collectResourceMetrics collects CPU, memory, and network metrics
func (pc *PrometheusCollector) collectResourceMetrics(ctx context.Context, service string, endpoint *models.Endpoint, timeRange models.TimeRange) (*models.ResourceMetrics, error) {
	rm := &models.ResourceMetrics{
		Timestamp: time.Now(),
	}

	// CPU usage query
	cpuQuery := fmt.Sprintf(`avg(rate(container_cpu_usage_seconds_total{service="%s"}[%s]))`,
		service, pc.config.QueryInterval)
	cpuResult, warnings, err := pc.queryRange(ctx, cpuQuery, timeRange)
	if err == nil && cpuResult != nil {
		if len(warnings) > 0 {
			pc.logger.Debugf("CPU query warnings: %v", warnings)
		}
		rm.CPUCores = pc.avgValue(cpuResult)
	}

	// Memory usage query
	memQuery := fmt.Sprintf(`avg(container_memory_usage_bytes{service="%s"})`, service)
	memResult, warnings, err := pc.queryRange(ctx, memQuery, timeRange)
	if err == nil && memResult != nil {
		if len(warnings) > 0 {
			pc.logger.Debugf("Memory query warnings: %v", warnings)
		}
		rm.MemoryMB = pc.avgValue(memResult) / (1024 * 1024) // Convert to MB
	}

	// Network in query
	netInQuery := fmt.Sprintf(`sum(rate(container_network_receive_bytes_total{service="%s"}[%s]))`,
		service, pc.config.QueryInterval)
	netInResult, warnings, err := pc.queryRange(ctx, netInQuery, timeRange)
	if err == nil && netInResult != nil {
		if len(warnings) > 0 {
			pc.logger.Debugf("Network in query warnings: %v", warnings)
		}
		rm.NetworkInMB = pc.avgValue(netInResult) / (1024 * 1024) // Convert to MB
	}

	// Network out query
	netOutQuery := fmt.Sprintf(`sum(rate(container_network_transmit_bytes_total{service="%s"}[%s]))`,
		service, pc.config.QueryInterval)
	netOutResult, warnings, err := pc.queryRange(ctx, netOutQuery, timeRange)
	if err == nil && netOutResult != nil {
		if len(warnings) > 0 {
			pc.logger.Debugf("Network out query warnings: %v", warnings)
		}
		rm.NetworkOutMB = pc.avgValue(netOutResult) / (1024 * 1024) // Convert to MB
	}

	return rm, nil
}

// collectPerformanceMetrics collects request rate, latency, and error metrics
func (pc *PrometheusCollector) collectPerformanceMetrics(ctx context.Context, service string, endpoint *models.Endpoint, timeRange models.TimeRange) (*models.PerformanceMetrics, error) {
	pm := &models.PerformanceMetrics{
		Timestamp: time.Now(),
	}

	// Request rate query
	rateQuery := fmt.Sprintf(`sum(rate(http_requests_total{service="%s",endpoint="%s"}[%s]))`,
		service, endpoint.Path, pc.config.QueryInterval)
	rateResult, _, err := pc.queryRange(ctx, rateQuery, timeRange)
	if err == nil && rateResult != nil {
		pm.RequestRate = pc.avgValue(rateResult)
	}

	// Error rate query
	errorQuery := fmt.Sprintf(`sum(rate(http_requests_total{service="%s",endpoint="%s",status=~"5.."}[%s]))`,
		service, endpoint.Path, pc.config.QueryInterval)
	errorResult, _, err := pc.queryRange(ctx, errorQuery, timeRange)
	if err == nil && errorResult != nil {
		pm.ErrorRate = pc.avgValue(errorResult)
	}

	// Latency metrics
	latencyQuery := fmt.Sprintf(`histogram_quantile(0.50, rate(http_request_duration_seconds_bucket{service="%s",endpoint="%s"}[%s]))`,
		service, endpoint.Path, pc.config.QueryInterval)
	p50Result, _, err := pc.queryRange(ctx, latencyQuery, timeRange)
	if err == nil && p50Result != nil {
		pm.LatencyP50 = time.Duration(pc.avgValue(p50Result) * float64(time.Second))
	}

	latencyP95Query := fmt.Sprintf(`histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{service="%s",endpoint="%s"}[%s]))`,
		service, endpoint.Path, pc.config.QueryInterval)
	p95Result, _, err := pc.queryRange(ctx, latencyP95Query, timeRange)
	if err == nil && p95Result != nil {
		pm.LatencyP95 = time.Duration(pc.avgValue(p95Result) * float64(time.Second))
	}

	latencyP99Query := fmt.Sprintf(`histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{service="%s",endpoint="%s"}[%s]))`,
		service, endpoint.Path, pc.config.QueryInterval)
	p99Result, _, err := pc.queryRange(ctx, latencyP99Query, timeRange)
	if err == nil && p99Result != nil {
		pm.LatencyP99 = time.Duration(pc.avgValue(p99Result) * float64(time.Second))
	}

	// Average latency
	avgQuery := fmt.Sprintf(`avg(rate(http_request_duration_seconds_sum{service="%s",endpoint="%s"}[%s]) / rate(http_request_duration_seconds_count{service="%s",endpoint="%s"}[%s]))`,
		service, endpoint.Path, pc.config.QueryInterval, service, endpoint.Path, pc.config.QueryInterval)
	avgResult, _, err := pc.queryRange(ctx, avgQuery, timeRange)
	if err == nil && avgResult != nil {
		pm.LatencyAvg = time.Duration(pc.avgValue(avgResult) * float64(time.Second))
	}

	return pm, nil
}

// queryRange executes a range query against Prometheus
func (pc *PrometheusCollector) queryRange(ctx context.Context, query string, timeRange models.TimeRange) (model.Value, []string, error) {
	r := v1.Range{
		Start: timeRange.Start,
		End:   timeRange.End,
		Step:  pc.config.QueryInterval,
	}

	result, warnings, err := pc.client.QueryRange(ctx, query, r)
	if err != nil {
		return nil, warnings, err
	}

	return result, warnings, nil
}

// avgValue calculates the average value from a Prometheus result
func (pc *PrometheusCollector) avgValue(value model.Value) float64 {
	if value == nil {
		return 0.0
	}

	matrix, ok := value.(model.Matrix)
	if !ok {
		return 0.0
	}

	if len(matrix) == 0 {
		return 0.0
	}

	sum := 0.0
	count := 0

	for _, stream := range matrix {
		for _, sample := range stream.Values {
			sum += float64(sample.Value)
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return sum / float64(count)
}

// aggregateServiceMetrics aggregates endpoint metrics to service level
func (pc *PrometheusCollector) aggregateServiceMetrics(endpoints map[string]*models.EndpointMetrics) *models.ResourceMetrics {
	aggregate := &models.ResourceMetrics{
		Timestamp: time.Now(),
	}

	for _, em := range endpoints {
		if em.Resource != nil {
			aggregate.CPUCores += em.Resource.CPUCores
			aggregate.MemoryMB += em.Resource.MemoryMB
			aggregate.NetworkInMB += em.Resource.NetworkInMB
			aggregate.NetworkOutMB += em.Resource.NetworkOutMB
			aggregate.DiskReadMB += em.Resource.DiskReadMB
			aggregate.DiskWriteMB += em.Resource.DiskWriteMB
		}
	}

	return aggregate
}
