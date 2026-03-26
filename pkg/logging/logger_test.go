// Package logging contains unit tests for the structured logger and custom
// Gin middleware to ensure logs and capabilities are correctly registered.
package logging

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegisterWithDiscovery(t *testing.T) {
	mu := sync.Mutex{}
	registered := false

	// Mock Registry server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/register" {
			var reg ServiceRegistration
			json.NewDecoder(r.Body).Decode(&reg)
			if reg.Name == "test-svc" {
				mu.Lock()
				registered = true
				mu.Unlock()
				w.WriteHeader(http.StatusOK)
			}
		} else if r.URL.Path == "/heartbeat/test-svc" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	Init("test-svc")
	RegisterWithDiscovery(srv.URL, ServiceRegistration{
		Name:         "test-svc",
		Endpoint:     "http://localhost:1234",
		Capabilities: []CapabilityRegistration{{Name: "test", Endpoints: []string{"/test"}}},
	})

	// Wait for background registration
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.True(t, registered)
	mu.Unlock()
}
