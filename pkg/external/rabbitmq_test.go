// Package external contains tests for the RabbitMQ client to verify connection
// handling and message publishing/consuming logic.
package external

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockBroker(t *testing.T) {
	broker := &MockBroker{}
	ctx := context.Background()

	err := broker.Publish(ctx, "test-queue", map[string]string{"foo": "bar"})
	assert.NoError(t, err)

	err = broker.Subscribe(ctx, "test-queue", func(ctx context.Context, payload []byte) error {
		return nil
	})
	assert.NoError(t, err)
}

func TestEmulatedBroker(t *testing.T) {
	broker := &EmulatedBroker{Endpoint: "http://localhost:5000"}
	ctx := context.Background()

	err := broker.Publish(ctx, "test-queue", map[string]string{"foo": "bar"})
	assert.NoError(t, err)

	err = broker.Subscribe(ctx, "test-queue", func(ctx context.Context, payload []byte) error {
		return nil
	})
	assert.NoError(t, err)
}
