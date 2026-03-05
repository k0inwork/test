package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pum-go/pkg/logging"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/register", func(c *gin.Context) {
		var info ServiceInfo
		if err := c.ShouldBindJSON(&info); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		info.LastUpdate = time.Now()

		mu.Lock()
		registry[info.Name] = &info
		mu.Unlock()
		c.JSON(http.StatusOK, gin.H{"status": "registered"})
	})

	r.GET("/services", func(c *gin.Context) {
		mu.RLock()
		defer mu.RUnlock()
		now := time.Now()
		active := make([]ServiceInfo, 0)
		for _, s := range registry {
			if now.Sub(s.LastUpdate) < 1*time.Second { // Short timeout for testing
				active = append(active, *s)
			}
		}
		c.JSON(http.StatusOK, active)
	})

	r.POST("/heartbeat/:name", func(c *gin.Context) {
		name := c.Param("name")
		mu.Lock()
		defer mu.Unlock()
		if s, ok := registry[name]; ok {
			s.LastUpdate = time.Now()
			c.JSON(http.StatusOK, gin.H{"status": "alive"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "service not registered"})
	})

	return r
}

func TestRegistry(t *testing.T) {
	router := setupRouter()

	t.Run("Register service", func(t *testing.T) {
		info := ServiceInfo{
			Name: "test-svc",
			Endpoint: "http://test:123",
			Capabilities: []logging.CapabilityRegistration{
				{Name: "test", Endpoints: []string{"/test"}},
			},
		}
		body, _ := json.Marshal(info)
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		req, _ = http.NewRequest("GET", "/services", nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		var services []ServiceInfo
		json.Unmarshal(resp.Body.Bytes(), &services)
		assert.Len(t, services, 1)
		assert.Equal(t, "test-svc", services[0].Name)
	})

	t.Run("Heartbeat updates timestamp", func(t *testing.T) {
		mu.RLock()
		oldUpdate := registry["test-svc"].LastUpdate
		mu.RUnlock()

		time.Sleep(10 * time.Millisecond)
		req, _ := http.NewRequest("POST", "/heartbeat/test-svc", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		mu.RLock()
		assert.True(t, registry["test-svc"].LastUpdate.After(oldUpdate))
		mu.RUnlock()
	})

	t.Run("Cleanup stale services", func(t *testing.T) {
		// Wait for timeout (we set 1s in setupRouter for test)
		time.Sleep(1100 * time.Millisecond)

		req, _ := http.NewRequest("GET", "/services", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		var services []ServiceInfo
		json.Unmarshal(resp.Body.Bytes(), &services)
		assert.Len(t, services, 0)
	})
}
