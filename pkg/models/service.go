package models

import "time"

// Service represents a microservice in the architecture
type Service struct {
	Name         string            `json:"name" yaml:"name"`
	Path         string            `json:"path" yaml:"path"`
	Endpoints    []*Endpoint       `json:"endpoints" yaml:"endpoints"`
	Dependencies []*Dependency     `json:"dependencies" yaml:"dependencies"`
	Metadata     map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Endpoint represents an API endpoint within a service
type Endpoint struct {
	Path          string           `json:"path" yaml:"path"`
	Method        string           `json:"method" yaml:"method"`
	Service       *Service         `json:"-" yaml:"-"`
	Dependencies  []*Dependency    `json:"dependencies" yaml:"dependencies"`
	DirectCost    float64          `json:"direct_cost" yaml:"direct_cost"`
	TotalCost     float64          `json:"total_cost" yaml:"total_cost"`
	CostBreakdown *CostBreakdown   `json:"cost_breakdown,omitempty" yaml:"cost_breakdown,omitempty"`
	Metrics       *EndpointMetrics `json:"metrics,omitempty" yaml:"metrics,omitempty"`
}

// Dependency represents a call from one service/endpoint to another
type Dependency struct {
	ID           string  `json:"id" yaml:"id"`
	FromService  string  `json:"from_service" yaml:"from_service"`
	FromEndpoint string  `json:"from_endpoint" yaml:"from_endpoint"`
	ToService    string  `json:"to_service" yaml:"to_service"`
	ToEndpoint   string  `json:"to_endpoint" yaml:"to_endpoint"`
	CallType     string  `json:"call_type" yaml:"call_type"` // http, grpc, internal
	Weight       float64 `json:"weight" yaml:"weight"`       // calls per parent call
	DetectedAt   string  `json:"detected_at" yaml:"detected_at"`
	LineNumber   int     `json:"line_number,omitempty" yaml:"line_number,omitempty"`
}

// CallGraph represents the complete dependency graph of all services
type CallGraph struct {
	Services     map[string]*Service `json:"services" yaml:"services"`
	Dependencies []*Dependency       `json:"dependencies" yaml:"dependencies"`
	GeneratedAt  time.Time           `json:"generated_at" yaml:"generated_at"`
	Metadata     map[string]string   `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// NewCallGraph creates a new empty call graph
func NewCallGraph() *CallGraph {
	return &CallGraph{
		Services:     make(map[string]*Service),
		Dependencies: make([]*Dependency, 0),
		GeneratedAt:  time.Now(),
		Metadata:     make(map[string]string),
	}
}

// AddService adds a service to the call graph
func (cg *CallGraph) AddService(service *Service) {
	if cg.Services == nil {
		cg.Services = make(map[string]*Service)
	}
	cg.Services[service.Name] = service
}

// AddDependency adds a dependency to the call graph
func (cg *CallGraph) AddDependency(dep *Dependency) {
	cg.Dependencies = append(cg.Dependencies, dep)
}

// GetService retrieves a service by name
func (cg *CallGraph) GetService(name string) (*Service, bool) {
	service, exists := cg.Services[name]
	return service, exists
}

// GetEndpoint retrieves an endpoint from a service
func (s *Service) GetEndpoint(path, method string) (*Endpoint, bool) {
	for _, ep := range s.Endpoints {
		if ep.Path == path && ep.Method == method {
			return ep, true
		}
	}
	return nil, false
}

// AddEndpoint adds an endpoint to the service
func (s *Service) AddEndpoint(endpoint *Endpoint) {
	endpoint.Service = s
	s.Endpoints = append(s.Endpoints, endpoint)
}
