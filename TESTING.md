# Testing Guide for MicroCost

## 🧪 Running Tests

### Quick Start

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test -v ./pkg/models

# Run specific test
go test -v -run TestNewCallGraph ./pkg/models
```

### Using Makefile

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run with race detector
make test-race

# Run benchmarks
make test-bench
```

## 📁 Test Structure

```
microcost/
├── pkg/
│   ├── models/
│   │   ├── service_test.go      # Service & CallGraph tests
│   │   ├── metrics_test.go      # Metrics models tests
│   │   └── cost_test.go         # Cost calculation tests
│   └── config/
│       └── config_test.go       # Configuration tests
├── internal/
│   └── graph/
│       └── graph_test.go        # Graph algorithms tests
└── Makefile                     # Test automation
```

## ✅ Test Coverage

### Models Package (`pkg/models`)

**service_test.go** - 17 tests
- `TestNewCallGraph` - CallGraph initialization
- `TestAddService` - Adding services to graph
- `TestAddDependency` - Adding dependencies
- `TestServiceAddEndpoint` - Endpoint management
- `TestGetEndpointNotFound` - Error handling
- `TestCallGraphMultipleServices` - Multiple service scenarios

**metrics_test.go** - 5 tests
- `TestNewMetricsSnapshot` - Metrics snapshot creation
- `TestAddServiceMetrics` - Adding service metrics
- `TestResourceMetrics` - Resource usage tracking
- `TestPerformanceMetrics` - Performance data validation
- `TestEndpointMetrics` - Per-endpoint metrics

**cost_test.go** - 6 tests
- `TestNewCostReport` - Cost report initialization
- `TestAddServiceCost` - Adding service costs
- `TestCalculateTotalCost` - Total cost calculation
- `TestNewCostBreakdown` - Cost breakdown generation
- `TestEndpointCost` - Endpoint cost validation
- `TestDownstreamCost` - Downstream cost attribution

### Config Package (`pkg/config`)

**config_test.go** - 8 tests
- `TestDefaultConfig` - Default configuration validation
- `TestValidate` - Configuration validation rules
- `TestValidateAutoCorrect` - Auto-correction behavior
- `TestLoadNonExistentFile` - Missing file handling
- `TestEnvironmentVariableOverride` - ENV variable support

### Graph Package (`internal/graph`)

**graph_test.go** - 16 tests
- `TestNewGraph` - Graph initialization
- `TestAddNode` - Node addition
- `TestAddNodeDuplicate` - Duplicate handling
- `TestAddEdge` - Edge creation
- `TestGetNode` - Node retrieval
- `TestGetOutgoingEdges` - Outgoing edges
- `TestGetIncomingEdges` - Incoming edges
- `TestHasCycleNoCycle` - Acyclic graph detection
- `TestHasCycleWithCycle` - Cycle detection
- `TestTopologicalSort` - Topological ordering
- `TestTopologicalSortWithCycle` - Cycle error handling
- `TestFindAllPaths` - Path finding algorithm
- `TestGetAllNodes` - Node enumeration
- `TestGetAllEdges` - Edge enumeration

## 📊 Test Coverage Report

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

View in browser:
```bash
# Windows
start coverage.html

# Linux/Mac
open coverage.html
```

## 🔍 Test Examples

### Testing CallGraph

```go
func TestCallGraphExample() {
    cg := models.NewCallGraph()
    
    // Add services
    service := &models.Service{
        Name: "payment-service",
        Endpoints: make([]*models.Endpoint, 0),
    }
    cg.AddService(service)
    
    // Verify
    assert.Equal(t, 1, len(cg.Services))
}
```

### Testing Cost Calculation

```go
func TestCostCalculation() {
    metrics := &models.ResourceMetrics{
        CPUCores: 2.0,
        MemoryMB: 1024.0,
    }
    
    model := &models.CostModel{
        CPUCostPerCoreHour: 0.05,
    }
    
    breakdown := models.NewCostBreakdown(metrics, nil, model, 1.0)
    
    // CPU cost: 2.0 * 0.05 * 1 hour = 0.10
    assert.Equal(t, 0.10, breakdown.CPUCost)
}
```

## 🎯 Test Best Practices

1. **Descriptive Names**: Use clear test function names
   - ✅ `TestAddServiceWithMultipleEndpoints`
   - ❌ `TestAdd`

2. **Table-Driven Tests**: For multiple scenarios
   ```go
   tests := []struct {
       name string
       input string
       want int
   }{
       {"empty", "", 0},
       {"single", "a", 1},
   }
   ```

3. **Setup & Teardown**: Use `defer` for cleanup
   ```go
   func TestWithCleanup(t *testing.T) {
       setup()
       defer teardown()
       // test code
   }
   ```

4. **Error Messages**: Include context
   ```go
   if got != want {
       t.Errorf("Expected %v, got %v", want, got)
   }
   ```

## 🚀 Continuous Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -html=coverage.out -o coverage.html
```

## 📈 Current Test Status

**Total Tests**: 52+  
**Packages Tested**: 5  
**Coverage Target**: 80%+  

**Status by Package**:
- ✅ `pkg/models` - 28 tests (Passing)
- ✅ `pkg/config` - 8 tests (Passing)
- ✅ `internal/graph` - 16 tests (Passing)

## 🔧 Running Specific Test Categories

```bash
# Unit tests only
go test -short ./...

# Integration tests
go test -run Integration ./...

# Race condition detection
go test -race ./...

# Benchmarks
go test -bench=. ./...
```

## 📝 Adding New Tests

1. Create `*_test.go` file in the same package
2. Import `"testing"`
3. Write functions starting with `Test`
4. Run `go test` to verify

Example:
```go
package mypackage

import "testing"

func TestMyFunction(t *testing.T) {
    result := MyFunction("input")
    expected := "output"
    
    if result != expected {
        t.Errorf("Expected %s, got %s", expected, result)
    }
}
```

## 🐛 Debugging Failed Tests

```bash
# Verbose output
go test -v ./pkg/models

# Stop on first failure
go test -failfast ./...

# Show test output even on success
go test -v -test.v ./...
```
