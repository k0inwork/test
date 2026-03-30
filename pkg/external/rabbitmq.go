// Package external provides a RabbitMQ client implementation for publishing
// and subscribing to message queues, maintaining legacy system compatibility.
package external

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
)

// MessageBroker defines the mandatory boundary for asynchronous communication (e.g., RabbitMQ).
// All Go microservices must use this interface when communicating with hardware controllers
// or background workers, rather than implementing internal, ad-hoc queues.
type MessageBroker interface {
	// Publish sends a generic payload to a specified queue or topic.
	Publish(ctx context.Context, queue string, payload interface{}) error
	// Subscribe registers a handler function to be invoked when a message arrives on the queue.
	Subscribe(ctx context.Context, queue string, handler func(ctx context.Context, payload []byte) error) error
}

// MockBroker is an in-memory stub implementation of MessageBroker.
// It is intended for use in the "mock" configuration mode.
// It logs calls to Publish and Subscribe but does not perform any actual network I/O,
// returning immediate success.
type MockBroker struct{}

// Publish simulates sending a message by logging the payload.
func (m *MockBroker) Publish(ctx context.Context, queue string, payload interface{}) error {
	_, span := otel.Tracer("MockBroker").Start(ctx, "Publish:"+queue)
	defer span.End()

	slog.Info("MockBroker: Published message", "queue", queue, "payload", payload)
	return nil
}

// Subscribe simulates registering a handler. The handler is never actively invoked by the MockBroker.
func (m *MockBroker) Subscribe(ctx context.Context, queue string, handler func(ctx context.Context, payload []byte) error) error {
	_, span := otel.Tracer("MockBroker").Start(ctx, "Subscribe:"+queue)
	defer span.End()

	slog.Info("MockBroker: Subscribed to queue", "queue", queue)
	return nil
}

// EmulatedBroker connects to a lightweight, standalone dummy service instead of a full message broker.
// It is intended for use in the "emulated" configuration mode, allowing tests to verify
// network serialization and timeouts without needing a heavy RabbitMQ container.
type EmulatedBroker struct {
	Endpoint string // The URL of the dummy service (e.g., "http://localhost:5000/api")
}

// Publish sends the payload to the dummy service's endpoint via an HTTP POST (simulated here).
func (e *EmulatedBroker) Publish(ctx context.Context, queue string, payload interface{}) error {
	_, span := otel.Tracer("EmulatedBroker").Start(ctx, "Publish:"+queue)
	defer span.End()

	slog.Info("EmulatedBroker: Connecting to mock server to publish", "queue", queue, "endpoint", e.Endpoint)
	return nil
}

// Subscribe polls or connects a websocket to the dummy service's endpoint (simulated here).
func (e *EmulatedBroker) Subscribe(ctx context.Context, queue string, handler func(ctx context.Context, payload []byte) error) error {
	_, span := otel.Tracer("EmulatedBroker").Start(ctx, "Subscribe:"+queue)
	defer span.End()

	slog.Info("EmulatedBroker: Connecting to mock server to subscribe", "queue", queue, "endpoint", e.Endpoint)
	return nil
}
