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
	resp, err := http.Get("http://127.0.0.1:16686")
	if err != nil {
		handleMissingEnv(t, "Jaeger UI not reachable")
	} else {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// 2. Verify Prometheus is reachable
	resp, err = http.Get("http://127.0.0.1:9090")
	if err != nil {
		handleMissingEnv(t, "Prometheus not reachable")
	} else {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// 3. Verify Jaeger Metrics endpoint (the one Prometheus scrapes)
	resp, err = http.Get("http://127.0.0.1:14269/metrics")
	if err != nil {
		handleMissingEnv(t, "Jaeger Metrics endpoint not reachable")
	} else {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// 4. Verify OTel Collector OTLP HTTP port is open
	resp, err = http.Get("http://127.0.0.1:4318")
	if err != nil {
		handleMissingEnv(t, "OTel Collector not reachable")
	} else {
		defer resp.Body.Close()
	}
}
