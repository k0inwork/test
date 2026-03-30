package backend_integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTelemetryEndToEnd(t *testing.T) {
	if os.Getenv("RUN_TELEMETRY_TESTS") != "true" {
		t.Skip("Skipping telemetry E2E test: RUN_TELEMETRY_TESTS not set to true")
	}

	// 1. Trigger an action that should generate a trace
	// Assuming the registry is running at localhost:8088
	resp, err := http.Get("http://localhost:8088/discovery")
	if err != nil {
		t.Fatalf("Failed to call registry: %v", err)
	}
	resp.Body.Close()

	// 2. Wait for the trace to propagate to Jaeger
	t.Log("Waiting for trace to propagate to Jaeger...")
	time.Sleep(5 * time.Second)

	// 3. Query Jaeger API for traces from the 'registry' service
	// Jaeger Query API is usually at :16686
	jaegerURL := "http://localhost:16686/api/traces?service=registry"

	var traces struct {
		Data []interface{} `json:"data"`
	}

	for i := 0; i < 5; i++ {
		resp, err = http.Get(jaegerURL)
		if err != nil {
			t.Logf("Jaeger API not ready yet (attempt %d/5)", i+1)
			time.Sleep(2 * time.Second)
			continue
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&traces); err != nil {
			t.Fatalf("Failed to decode Jaeger response: %v", err)
		}

		if len(traces.Data) > 0 {
			t.Logf("Successfully found %d traces in Jaeger for 'registry'", len(traces.Data))
			return
		}

		t.Logf("No traces found for 'registry' yet (attempt %d/5)", i+1)
		time.Sleep(2 * time.Second)
	}

	assert.NotEmpty(t, traces.Data, "Expected to find traces in Jaeger for the 'registry' service")
}
