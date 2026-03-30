prometheus: .bin/prometheus --config.file=prometheus.yml
otelcol: .bin/otelcol-contrib --config=otelcol-config.yaml
jaeger: COLLECTOR_OTLP_GRPC_HOST_PORT=:14317 COLLECTOR_OTLP_HTTP_HOST_PORT=:14318 METRICS_STORAGE_TYPE=prometheus .bin/jaeger-all-in-one --prometheus.server-url=http://localhost:9090 --prometheus.query.normalize-calls=true --prometheus.query.normalize-duration=true --prometheus.query.support-spanmetrics-connector=true --query.ui-config=jaeger-ui.json
registry: OTEL_EXPORTER_OTLP_INSECURE=true go run services/registry/main.go
identity: OTEL_EXPORTER_OTLP_INSECURE=true go run services/identity/main.go
product: OTEL_EXPORTER_OTLP_INSECURE=true go run services/product/main.go
inventory: OTEL_EXPORTER_OTLP_INSECURE=true go run services/inventory/main.go
network: OTEL_EXPORTER_OTLP_INSECURE=true go run services/network/main.go
task: OTEL_EXPORTER_OTLP_INSECURE=true go run services/task/main.go
external-modules: OTEL_EXPORTER_OTLP_INSECURE=true go run services/external-modules/main.go
terminal: OTEL_EXPORTER_OTLP_INSECURE=true go run services/terminal/main.go
external-data: OTEL_EXPORTER_OTLP_INSECURE=true go run services/external-data/main.go
frontend: OTEL_EXPORTER_OTLP_INSECURE=true go run services/frontend/main.go
compatibility: OTEL_EXPORTER_OTLP_INSECURE=true go run services/compatibility/main.go
gws: OTEL_EXPORTER_OTLP_INSECURE=true go run services/gws/main.go
keyservice: OTEL_EXPORTER_OTLP_INSECURE=true go run services/keyservice/main.go
