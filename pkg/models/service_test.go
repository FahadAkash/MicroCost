package models

import (
	"testing"
)

func TestNewCallGraph(t *testing.T) {
	cg := NewCallGraph()

	if cg == nil {
		t.Fatal("NewCallGraph returned nil")
	}

	if cg.Services == nil {
		t.Error("Services map not initialized")
	}

	if cg.Dependencies == nil {
		t.Error("Dependencies slice not initialized")
	}

	if cg.GeneratedAt.IsZero() {
		t.Error("GeneratedAt not set")
	}
}

func TestAddService(t *testing.T) {
	cg := NewCallGraph()

	service := &Service{
		Name:         "test-service",
		Path:         "/test/path",
		Endpoints:    make([]*Endpoint, 0),
		Dependencies: make([]*Dependency, 0),
	}

	cg.AddService(service)

	if len(cg.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(cg.Services))
	}

	retrieved, exists := cg.GetService("test-service")
	if !exists {
		t.Error("Service not found after adding")
	}

	if retrieved.Name != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", retrieved.Name)
	}
}

func TestAddDependency(t *testing.T) {
	cg := NewCallGraph()

	dep := &Dependency{
		ID:          "dep1",
		FromService: "service-a",
		ToService:   "service-b",
		CallType:    "http",
		Weight:      1.0,
	}

	cg.AddDependency(dep)

	if len(cg.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(cg.Dependencies))
	}

	if cg.Dependencies[0].ID != "dep1" {
		t.Errorf("Expected dependency ID 'dep1', got '%s'", cg.Dependencies[0].ID)
	}
}

func TestServiceAddEndpoint(t *testing.T) {
	service := &Service{
		Name:      "test-service",
		Path:      "/test",
		Endpoints: make([]*Endpoint, 0),
	}

	endpoint := &Endpoint{
		Path:   "/api/users",
		Method: "GET",
	}

	service.AddEndpoint(endpoint)

	if len(service.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(service.Endpoints))
	}

	if endpoint.Service != service {
		t.Error("Endpoint service reference not set")
	}

	// Test GetEndpoint
	retrieved, exists := service.GetEndpoint("/api/users", "GET")
	if !exists {
		t.Error("Endpoint not found")
	}

	if retrieved.Path != "/api/users" {
		t.Errorf("Expected path '/api/users', got '%s'", retrieved.Path)
	}
}

func TestGetEndpointNotFound(t *testing.T) {
	service := &Service{
		Name:      "test-service",
		Endpoints: make([]*Endpoint, 0),
	}

	_, exists := service.GetEndpoint("/nonexistent", "GET")
	if exists {
		t.Error("Expected endpoint to not exist")
	}
}

func TestCallGraphMultipleServices(t *testing.T) {
	cg := NewCallGraph()

	services := []*Service{
		{Name: "service-1", Path: "/path1", Endpoints: make([]*Endpoint, 0), Dependencies: make([]*Dependency, 0)},
		{Name: "service-2", Path: "/path2", Endpoints: make([]*Endpoint, 0), Dependencies: make([]*Dependency, 0)},
		{Name: "service-3", Path: "/path3", Endpoints: make([]*Endpoint, 0), Dependencies: make([]*Dependency, 0)},
	}

	for _, service := range services {
		cg.AddService(service)
	}

	if len(cg.Services) != 3 {
		t.Errorf("Expected 3 services, got %d", len(cg.Services))
	}

	for _, service := range services {
		retrieved, exists := cg.GetService(service.Name)
		if !exists {
			t.Errorf("Service %s not found", service.Name)
		}
		if retrieved.Name != service.Name {
			t.Errorf("Expected %s, got %s", service.Name, retrieved.Name)
		}
	}
}
