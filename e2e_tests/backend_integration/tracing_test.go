package backend_integration

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracingConnections(t *testing.T) {
	if os.Getenv("RUN_TELEMETRY_TESTS") != "true" {
		t.Skip("Skipping telemetry test: RUN_TELEMETRY_TESTS not set to true")
	}

	// 1. Verify Jaeger UI is reachable
	resp, err := http.Get("http://localhost:16686")
	if err != nil {
		t.Skip("Jaeger UI not reachable")
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 2. Verify Prometheus is reachable
	resp, err = http.Get("http://localhost:9090")
	if err != nil {
		t.Skip("Prometheus not reachable")
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 3. Verify Jaeger Metrics endpoint (the one Prometheus scrapes)
	// Default Jaeger admin port is 14269
	resp, err = http.Get("http://localhost:14269/metrics")
	if err != nil {
		t.Skip("Jaeger Metrics endpoint not reachable")
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// 4. Verify OTel Collector OTLP HTTP port is open
	resp, err = http.Get("http://localhost:4318")
	if err != nil {
		t.Skip("OTel Collector not reachable")
	}
	defer resp.Body.Close()
}
