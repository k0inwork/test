prometheus: .bin/prometheus --config.file=prometheus.yml
otelcol: .bin/otelcol-contrib --config=otelcol-config.yaml
jaeger: COLLECTOR_OTLP_GRPC_HOST_PORT=:14317 COLLECTOR_OTLP_HTTP_HOST_PORT=:14318 METRICS_STORAGE_TYPE=prometheus .bin/jaeger-all-in-one --prometheus.server-url=http://localhost:9090 --prometheus.query.normalize-calls=true --prometheus.query.normalize-duration=true --prometheus.query.support-spanmetrics-connector=true --query.ui-config=jaeger-ui.json
registry: go run services/registry/main.go
identity: go run services/identity/main.go
product: go run services/product/main.go
inventory: go run services/inventory/main.go
network: go run services/network/main.go
task: go run services/task/main.go
external-modules: go run services/external-modules/main.go
terminal: go run services/terminal/main.go
external-data: go run services/external-data/main.go
frontend: go run services/frontend/main.go
compatibility: go run services/compatibility/main.go
gws: go run services/gws/main.go
keyservice: go run services/keyservice/main.go
