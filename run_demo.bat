@echo off
echo ==========================================
echo      MicroCost Demo Environment
echo ==========================================

echo [1/3] Building MicroCost...
go build -o microcost.exe main.go
if %errorlevel% neq 0 (
    echo Build failed!
    exit /b %errorlevel%
)

echo [2/3] Analyzing Demo Services...
.\microcost.exe analyze --config demo/config.yaml
if %errorlevel% neq 0 (
    echo Analysis failed!
    exit /b %errorlevel%
)

echo [3/3] Demo Complete! 
echo Check demo/reports/ for JSON output.
echo.
echo NOTE: To see runtime metrics, you need to run:
echo    1. Prometheus (docker run -p 9090:9090 prom/prometheus)
echo    2. go run demo/product-service/main.go
echo    3. go run demo/pricing-service/main.go
echo.
pause
