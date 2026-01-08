# MicroCost 💰

> **Microservices Dependency & Cost Mapper** - A production-ready CLI tool that maps your microservices architecture and calculates the true cost of each API endpoint by tracing dependencies across service boundaries.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## 🚀 Features

- **📊 Static Code Analysis** - Automatically scans Go codebases to discover services, HTTP handlers, and gRPC methods
- **🔍 Dependency Detection** - Identifies HTTP and gRPC calls to build complete service dependency graphs
- **📈 Metrics Collection** - Pulls CPU, memory, network, latency, and request metrics from Prometheus
- **💰 Cost Attribution** - Calculates true endpoint costs including all downstream service costs
- **🎨 Rich Visualization** - ASCII trees, tables, and JSON/YAML exports
- **⚡ Performance Analysis** - Identifies bottlenecks and cost leaks in your call chains
- **🔧 Production Ready** - Comprehensive logging, error handling, and configuration management

## 🏛️ Enterprise Benefits

For large organizations with complex microservices architectures, **MicroCost** provides:

- **Accurate Cost Attribution** - Reveal the true drivers of cloud spend by tracing costs through the call graph. See how upstream services impact your downstream resource usage.
- **Zombie Service Detection** - Identify expensive, orphaned services that have zero active dependents and can be safely decommissioned.
- **Developer-Led FinOps** - Move cost awareness "left." Let engineers see the financial impact of architectural changes directly in their development environment or CI/CD pipeline.
- **Cost Leak Detection** - Catch expensive API loops or unoptimized gRPC calls in real-time by combining static dependency maps with runtime Prometheus metrics.
- **Architectural Decision Support** - Use data-driven cost analysis to decide whether to split a monolith or consolidate microservices based on projected infrastructure overhead.

## 📦 Installation

### From Source

```bash
cd microcost
go mod download
go build -o microcost main.go
```

### Quick Test

```bash
./microcost --version
./microcost --help
```

## 🎯 Quick Start

### 1. Configure

Create a `config.yaml`:

```yaml
analysis:
  paths:
    - ./services
  
prometheus:
  url: "http://localhost:9090"

cost_model:
  provider: "aws"
  cpu_cost_per_core_hour: 0.0416
  memory_cost_per_gb_hour: 0.0052
```

### 2. Run Complete Pipeline

```bash
# Analyze, collect metrics, and calculate costs in one command
./microcost all --duration 1h --output ./reports
```

### 3. Or Run Step-by-Step

```bash
# Step 1: Analyze code and build dependency graph
./microcost analyze --paths ./services --output callgraph.json

# Step 2: Collect metrics from Prometheus
./microcost collect --callgraph callgraph.json --duration 1h --output metrics.json

# Step 3: Calculate costs
./microcost calculate --callgraph callgraph.json --metrics metrics.json --output cost-report.json
```

## 📖 Usage

### Analyze Command

Scan codebase to discover services and dependencies:

```bash
./microcost analyze \
  --paths ./services,./api \
  --output callgraph.json \
  --visualize
```

**Options:**
- `--paths, -p` - Comma-separated paths to analyze (default: `.`)
- `--output, -o` - Output file path (default: `callgraph.json`)
- `--format, -f` - Output format: `json`, `yaml` (default: `json`)
- `--visualize, -v` - Show ASCII dependency tree (default: `true`)

### Collect Command

Gather runtime metrics from Prometheus:

```bash
./microcost collect \
  --callgraph callgraph.json \
  --duration 24h \
  --output metrics.json
```

**Options:**
- `--callgraph, -g` - Call graph input file (default: `callgraph.json`)
- `--duration, -d` - Time window for metrics: `1h`, `24h`, `7d` (default: `1h`)
- `--output, -o` - Output file path (default: `metrics.json`)

### Calculate Command

Calculate endpoint costs with downstream attribution:

```bash
./microcost calculate \
  --callgraph callgraph.json \
  --metrics metrics.json \
  --output cost-report.json \
  --visualize
```

**Options:**
- `--callgraph, -g` - Call graph input file
- `--metrics, -m` - Metrics input file
- `--output, -o` - Output file path
- `--format, -f` - Output format: `json`, `yaml`, `ascii`
- `--visualize, -v` - Show ASCII cost report

### All Command

Run complete pipeline:

```bash
./microcost all --duration 6h --output ./reports
```

## 📊 Example Output

### ASCII Cost Report

```
=== MICROSERVICES COST REPORT ===

Total Cost: $125.4560
Time Range: 2024-01-08 09:00 to 2024-01-08 15:00
Services: 8
Provider: aws (us-east-1)

▶ Top Costly Endpoints

+------+-----------+-------------+-------------+------------+------------+-----------+
| Rank | Service   | Endpoint    | Direct Cost | Downstream | Total Cost | $/Request |
+------+-----------+-------------+-------------+------------+------------+-----------+
|    1 | checkout  | /checkout   | $0.0200     | $0.1300    | $0.1500    | $0.000150 |
|    2 | inventory | /check      | $0.0180     | $0.0450    | $0.0630    | $0.000063 |
+------+-----------+-------------+-------------+------------+------------+-----------+

💡 Recommendations

1. Consider optimizing checkout/checkout - highest cost endpoint at $0.0002 per request
2. inventory/check has 5x more downstream cost than direct cost - review dependency chain
```

### Dependency Tree

```
▶ checkout
  ├─ /payment (http, weight: 1.0)
  │  └─ payment
  ├─ /inventory (http, weight: 1.0)
  │  └─ inventory
  │     ├─ /warehouse (http, weight: 1.0)
  │     └─ /supplier (http, weight: 0.5)
  └─ /notification (http, weight: 1.0)
     └─ notification
```

## ⚙️ Configuration

See [`config.yaml`](config.yaml) for full configuration reference.

### Key Settings

**Analysis:**
- Service discovery paths
- Exclude patterns
- Service naming patterns

**Prometheus:**
- Server URL and credentials
- Query intervals
- Custom PromQL queries

**Cost Model:**
- Cloud provider (AWS, GCP, Azure)
- Per-resource pricing
- Custom cost models

**Output:**
- Default formats
- Color settings
- Top-N filtering

## 🏗️ Architecture

```
microcost/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command & config
│   ├── analyze.go         # Code analysis
│   ├── collect.go         # Metrics collection
│   ├── calculate.go       # Cost calculation
│   └── all.go             # Full pipeline
├── internal/
│   ├── analyzer/          # Static code analysis
│   │   ├── scanner.go     # AST scanner
│   │   ├── http_detector.go   # HTTP call detection
│   │   ├── grpc_detector.go   # gRPC call detection
│   │   └── graph_builder.go   # Dependency graph builder
│   ├── collector/         # Metrics collection
│   │   └── prometheus.go  # Prometheus client
│   ├── costengine/        # Cost calculation
│   │   └── calculator.go  # Cost attribution engine
│   ├── graph/             # Graph algorithms
│   │   └── graph.go       # Graph data structure
│   └── visualizer/        # Output generation
│       ├── ascii.go       # ASCII renderer
│       └── export.go      # JSON/YAML export
└── pkg/
    ├── config/            # Configuration management
    └── models/            # Data models
        ├── service.go     # Service & dependency models
        ├── metrics.go     # Metrics models
        └── cost.go        # Cost models
```

## 🔧 Development

### Build

```bash
go build -o microcost main.go
```

### Run Tests

```bash
go test ./...
```

### Run with Coverage

```bash
go test -cover ./...
```

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration
- [Prometheus Client](https://github.com/prometheus/client_golang) - Metrics
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Logrus](https://github.com/sirupsen/logrus) - Logging

## 📧 Support

For issues and questions, please open an issue on GitHub.

---

**Made with ❤️ for microservices observability**
