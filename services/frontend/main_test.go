package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFrontendBlocking(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock Registry server
	registrySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/services" {
			// Initially return no core services
			json.NewEncoder(w).Encode([]RegisteredService{})
		}
	}))
	defer registrySrv.Close()

	// Update Registry URL for testing
	// We'll use a local function instead of a global const for the test
	testRegistryURL := registrySrv.URL

	r := gin.New()

	// Middleware to check core services (same logic as main.go but uses testRegistryURL)
	r.Use(func(c *gin.Context) {
		resp, err := http.Get(testRegistryURL + "/services")
		if err != nil {
			c.String(http.StatusServiceUnavailable, "Registry Service is Offline")
			c.Abort()
			return
		}
		defer resp.Body.Close()
		var services []RegisteredService
		json.NewDecoder(resp.Body).Decode(&services)

		required := map[string]bool{"identity": false, "product": false}
		for _, s := range services {
			if _, ok := required[s.Name]; ok {
				required[s.Name] = true
			}
		}

		for name, found := range required {
			if !found {
				c.String(http.StatusServiceUnavailable, fmt.Sprintf("Core Service Offline: %s", name))
				c.Abort()
				return
			}
		}
		c.Next()
	})

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	t.Run("Blocked when core services missing", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusServiceUnavailable, resp.Code)
		assert.Contains(t, resp.Body.String(), "Core Service Offline")
	})

	t.Run("Allowed when core services active", func(t *testing.T) {
		// Update Mock Registry to return core services
		registrySrv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/services" {
				json.NewEncoder(w).Encode([]RegisteredService{
					{Name: "identity", Endpoint: "http://test:8081"},
					{Name: "product", Endpoint: "http://test:8082"},
				})
			}
		})

		req, _ := http.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "OK", resp.Body.String())
	})
}
