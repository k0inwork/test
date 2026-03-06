#!/bin/bash

# Function to kill processes on exit
cleanup() {
    echo "Stopping all services..."
    kill $(jobs -p)
    exit
}

trap cleanup SIGINT SIGTERM

echo "Starting PUM Microservices (Go Registry-based Architecture)..."

echo "Setting up Prometheus for Monitoring..."
bash scripts/download_prometheus.sh
if [ -f "bin/prometheus" ]; then
    echo "Starting Prometheus on port 9090..."
    ./bin/prometheus --config.file=prometheus.yml > prometheus_output.log 2>&1 &
    sleep 2
else
    echo "Warning: Prometheus binary not found, monitoring data will not be scraped."
fi

echo "Setting up Jaeger for Distributed Tracing..."
bash scripts/download_jaeger.sh
if [ -f "bin/jaeger-all-in-one" ]; then
    echo "Starting Jaeger all-in-one on ports 16686 (UI) and 4318 (OTLP HTTP)..."
    # To connect Jaeger UI to a Prometheus-compatible metrics backend (e.g. for SPM), we set METRICS_STORAGE_TYPE=prometheus
    # By default, --prometheus.server-url is http://localhost:9090. If you are running VictoriaMetrics or Prometheus, make sure it is available there.
    METRICS_STORAGE_TYPE=prometheus ./bin/jaeger-all-in-one --prometheus.server-url=http://localhost:9090 > jaeger_output.log 2>&1 &
    sleep 2
else
    echo "Warning: Jaeger binary not found, tracing data will be dropped."
fi

# 1. Start Registry FIRST
echo "Starting Registry Service on :8088..."
go run services/registry/main.go &
sleep 2

# 2. Start Core Services
echo "Starting Identity Service on :8081..."
go run services/identity/main.go &

echo "Starting Product Service on :8082..."
go run services/product/main.go &

echo "Starting Inventory Service on :8083..."
go run services/inventory/main.go &

echo "Starting Network Service on :8084..."
go run services/network/main.go &

echo "Starting Task Service on :8085..."
go run services/task/main.go &

echo "Starting External Modules Proxy on :8086..."
go run services/external-modules/main.go &

echo "Starting External Data Service on :8089..."
go run services/external-data/main.go &

echo "Starting Terminal Service on :8087..."
go run services/terminal/main.go &

# Wait for services to initialize and register
sleep 5

# 4. Start Frontend
# Wait for services to initialize

echo "Starting Frontend Service on :8080..."
go run services/frontend/main.go &

echo "------------------------------------------------"
echo "Nodes List: http://localhost:8080/nodes"
echo "User List: http://localhost:8080/users"
echo "Registry: http://localhost:8088/services"
echo "All systems running!"
echo "Dashboard: http://localhost:8080"

echo "------------------------------------------------"

# Keep script running
wait
