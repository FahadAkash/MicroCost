# MicroCost Demo Environment

This folder contains a minimal microservices setup to demonstrate **MicroCost's** capabilities.

## 📂 Structure

- **product-service/**: A service running on `:8081`. Calls the pricing service.
- **pricing-service/**: A service running on `:8082`.
- **config.yaml**: Configuration to scan this specific folder.

## 🚀 How to Run

### 1. Quick Static Analysis
Double-click `run_demo.bat` in the root folder, or run:

```bash
./run_demo.bat
```

This will compile MicroCost and scan the `demo/` folder, outputting the dependency graph (Product -> Pricing).

### 2. Full "Real-World" Simulation
To see metrics and costs in action:

1. **Start Prometheus** (Docker required):
   ```bash
   docker run -p 9090:9090 prom/prometheus
   ```
   *Note: configure Prometheus to scrape localhost:8081 and localhost:8082*

2. **Start the Services**:
   Open two terminals:
   ```bash
   # Terminal 1
   go run demo/product-service/main.go
   
   # Terminal 2
   go run demo/pricing-service/main.go
   ```

3. **Generate Traffic**:
   Data will start flowing.
   ```bash
   curl http://localhost:8081/product
   ```

4. **Run MicroCost**:
   ```bash
   ./microcost all --config demo/config.yaml
   ```

## 🔒 Git Ignore
This `demo/` folder is added to `.gitignore`, so you can mess around with it without affecting the main repository structure.
