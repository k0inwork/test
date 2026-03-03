package external

import (
	"testing"
)

func TestMockBroker(t *testing.T) {
	broker := &MockBroker{}

	// Test Publish
	err := broker.Publish("test_queue", map[string]string{"msg": "hello"})
	if err != nil {
		t.Errorf("MockBroker.Publish returned an error: %v", err)
	}

	// Test Subscribe
	handlerCalled := false
	handler := func(payload []byte) error {
		handlerCalled = true
		return nil
	}

	err = broker.Subscribe("test_queue", handler)
	if err != nil {
		t.Errorf("MockBroker.Subscribe returned an error: %v", err)
	}
	// Note: In a pure mock, the handler isn't actively called by a background process,
	// but the registration should succeed.
	if handlerCalled {
		t.Log("Handler was called synchronously (unexpected for pure MockBroker but fine if implemented that way)")
	}
}

func TestEmulatedBroker(t *testing.T) {
	broker := &EmulatedBroker{Endpoint: "http://localhost:9999/dummy"}

	// Test Publish
	err := broker.Publish("test_queue", map[string]string{"msg": "hello"})
	if err != nil {
		t.Errorf("EmulatedBroker.Publish returned an error: %v", err)
	}

	// Test Subscribe
	handler := func(payload []byte) error {
		return nil
	}
	err = broker.Subscribe("test_queue", handler)
	if err != nil {
		t.Errorf("EmulatedBroker.Subscribe returned an error: %v", err)
	}
}
