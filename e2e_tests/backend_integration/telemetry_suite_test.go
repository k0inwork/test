package backend_integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"pum-go/pkg/tracing"
)

// TestInitTracer_GlobalSideEffects verifies that the InitTracer function sets global state properly
// and that this state can be cleared. It essentially functions as a broader suite integration test.
func TestInitTracer_GlobalSideEffects(t *testing.T) {
	// 1. Initial State Setup
	originalTracerProvider := otel.GetTracerProvider()
	originalPropagator := otel.GetTextMapPropagator()

	defer func() {
		otel.SetTracerProvider(originalTracerProvider)
		otel.SetTextMapPropagator(originalPropagator)
	}()

	// 2. Execution
	tp, err := tracing.InitTracer("telemetry-integration-service")

	// 3. Validation
	require.NoError(t, err)
	require.NotNil(t, tp)

	defer func() {
		err := tp.Shutdown(context.Background())
		assert.NoError(t, err)
	}()

	assert.Equal(t, tp, otel.GetTracerProvider())
}
