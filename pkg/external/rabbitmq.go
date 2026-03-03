package external

import (
	"log/slog"
)

// MessageBroker represents the interface for async communication (e.g. RabbitMQ)
type MessageBroker interface {
	Publish(queue string, payload interface{}) error
	Subscribe(queue string, handler func(payload []byte) error) error
}

// MockBroker is an in-memory mock implementation of MessageBroker
type MockBroker struct{}

func (m *MockBroker) Publish(queue string, payload interface{}) error {
	slog.Info("MockBroker: Published message", "queue", queue, "payload", payload)
	return nil
}

func (m *MockBroker) Subscribe(queue string, handler func(payload []byte) error) error {
	slog.Info("MockBroker: Subscribed to queue", "queue", queue)
	return nil
}

// EmulatedBroker connects to a local test HTTP or lightweight dummy service
// representing the external broker bounds without running full RabbitMQ
type EmulatedBroker struct {
	Endpoint string
}

func (e *EmulatedBroker) Publish(queue string, payload interface{}) error {
	slog.Info("EmulatedBroker: Connecting to mock server to publish", "queue", queue, "endpoint", e.Endpoint)
	// Example: would do http.Post(e.Endpoint+"/publish", ...)
	return nil
}

func (e *EmulatedBroker) Subscribe(queue string, handler func(payload []byte) error) error {
	slog.Info("EmulatedBroker: Connecting to mock server to subscribe", "queue", queue, "endpoint", e.Endpoint)
	return nil
}
