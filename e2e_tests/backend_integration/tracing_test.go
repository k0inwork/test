package backend_integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracingConnections(t *testing.T) {
	resp, err := http.Get("http://localhost:16686")
	if err != nil {
		handleMissingEnv(t, "Jaeger UI not reachable")
	} else {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	resp, err = http.Get("http://localhost:9090")
	if err != nil {
		handleMissingEnv(t, "Prometheus not reachable")
	} else {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	resp, err = http.Get("http://localhost:14269/metrics")
	if err != nil {
		handleMissingEnv(t, "Jaeger Metrics endpoint not reachable")
	} else {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	resp, err = http.Get("http://localhost:4318")
	if err != nil {
		handleMissingEnv(t, "OTel Collector not reachable")
	} else {
		defer resp.Body.Close()
	}
}
