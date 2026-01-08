# Test Makefile for MicroCost

.PHONY: test test-unit test-integration test-coverage test-race clean help

# Run all tests
test:
	@echo "Running all tests..."
	go test -v ./...

# Run only unit tests
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -run Integration ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -v -race ./...

# Run benchmarks
test-bench:
	@echo "Running benchmarks..."
	go test -v -bench=. -benchmem ./...

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	rm -f coverage.out coverage.html
	rm -f *.test
	go clean -testcache

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Run all quality checks
check: fmt vet lint test

# Build the application
build:
	@echo "Building application..."
	go build -o microcost.exe main.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Show help
help:
	@echo "MicroCost Test Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  test          - Run all tests"
	@echo "  test-unit     - Run unit tests only"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-race     - Run tests with race detector"
	@echo "  test-bench    - Run benchmarks"
	@echo "  clean         - Clean test artifacts"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  check         - Run all quality checks"
	@echo "  build         - Build the application"
	@echo "  deps          - Install dependencies"
	@echo "  help          - Show this help message"
