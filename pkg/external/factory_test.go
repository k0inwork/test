// Package external contains tests for the factory implementation to ensure
// proper instantiation of external integration providers.
package external

import (
	"testing"
)

func TestNewMessageBroker(t *testing.T) {
	tests := []struct {
		name         string
		mode         string
		endpoint     string
		realEndpoint string
		wantType     string
	}{
		{
			name:         "Creates MockBroker for 'mock' mode",
			mode:         "mock",
			endpoint:     "",
			realEndpoint: "",
			wantType:     "*external.MockBroker",
		},
		{
			name:         "Creates EmulatedBroker for 'emulated' mode",
			mode:         "emulated",
			endpoint:     "http://test",
			realEndpoint: "",
			wantType:     "*external.EmulatedBroker",
		},
		{
			name:         "Falls back to MockBroker for 'real' mode currently (Phase 1)",
			mode:         "real",
			endpoint:     "amqp://test",
			realEndpoint: "amqps://prod",
			wantType:     "*external.MockBroker",
		},
		{
			name:         "Defaults to MockBroker for unknown mode",
			mode:         "unknown_mode",
			endpoint:     "",
			realEndpoint: "",
			wantType:     "*external.MockBroker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broker := NewMessageBroker(tt.mode, tt.endpoint, tt.realEndpoint)
			if broker == nil {
				t.Fatalf("NewMessageBroker returned nil")
			}

			// A simple type assertion string representation check
			gotType := ""
			switch broker.(type) {
			case *MockBroker:
				gotType = "*external.MockBroker"
			case *EmulatedBroker:
				gotType = "*external.EmulatedBroker"
			default:
				gotType = "unknown"
			}

			if gotType != tt.wantType {
				t.Errorf("NewMessageBroker() = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}
