package main

import (
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ServiceInfo struct {
	Name         string    `json:"name"`
	Endpoint     string    `json:"endpoint"`
	Capabilities []string  `json:"capabilities"`
	IsCore       bool      `json:"is_core"`
	LastUpdate   time.Time `json:"last_update"`
}

var (
	registry = make(map[string]*ServiceInfo)
	mu       sync.RWMutex
)

func main() {
	logging.Init("registry")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	// Register a service
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

		slog.Info("Service registered", "name", info.Name, "endpoint", info.Endpoint)
		c.JSON(http.StatusOK, gin.H{"status": "registered"})
	})

	// Get all registered services
	r.GET("/services", func(c *gin.Context) {
		mu.RLock()
		defer mu.RUnlock()

		// Cleanup stale services
		now := time.Now()
		active := make([]ServiceInfo, 0)
		for _, s := range registry {
			if now.Sub(s.LastUpdate) < 60*time.Second {
				active = append(active, *s)
			}
		}

		c.JSON(http.StatusOK, active)
	})

	// Heartbeat
	r.POST("/heartbeat/:name", func(c *gin.Context) {
		name := c.Param("name")
		mu.Lock()
		if s, ok := registry[name]; ok {
			s.LastUpdate = time.Now()
			mu.Unlock()
			c.JSON(http.StatusOK, gin.H{"status": "alive"})
			return
		}
		mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "service not registered"})
	})

	slog.Info("Registry service starting", "port", 8088)
	r.Run(":8088")
}
