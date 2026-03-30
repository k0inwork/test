package tracing

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func TestInitTracer_SuccessWithDefaults(t *testing.T) {
	// Isolate env state
	os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Unsetenv("OTEL_EXPORTER_OTLP_INSECURE")

	originalTP := otel.GetTracerProvider()
	originalProp := otel.GetTextMapPropagator()
	defer func() {
		otel.SetTracerProvider(originalTP)
		otel.SetTextMapPropagator(originalProp)
	}()

	tp, err := InitTracer("test-service")
	require.NoError(t, err)
	require.NotNil(t, tp)

	defer func() {
		err := tp.Shutdown(context.Background())
		assert.NoError(t, err)
	}()

	propagator := otel.GetTextMapPropagator()
	assert.ElementsMatch(t, []string{"traceparent", "tracestate", "baggage"}, propagator.Fields())
}

func TestInitTracer_SuccessWithCustomEnv(t *testing.T) {
	// Set custom environment variables
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "custom-host:4318")
	t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")

	originalTP := otel.GetTracerProvider()
	originalProp := otel.GetTextMapPropagator()
	defer func() {
		otel.SetTracerProvider(originalTP)
		otel.SetTextMapPropagator(originalProp)
	}()

	tp, err := InitTracer("test-service")
	require.NoError(t, err)
	require.NotNil(t, tp)

	defer func() {
		err := tp.Shutdown(context.Background())
		assert.NoError(t, err)
	}()
}
