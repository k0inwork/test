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
	Enabled      bool      `json:"enabled"`
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

	r.POST("/register", func(c *gin.Context) {
		var info ServiceInfo
		if err := c.ShouldBindJSON(&info); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		info.LastUpdate = time.Now()
		mu.Lock()
		if existing, ok := registry[info.Name]; ok {
			info.Enabled = existing.Enabled
		} else {
			info.Enabled = true
		}
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
			if now.Sub(s.LastUpdate) < 60*time.Second && s.Enabled {
				active = append(active, *s)
			}
		}
		c.JSON(http.StatusOK, active)
	})

	r.GET("/admin/services", func(c *gin.Context) {
		mu.RLock()
		defer mu.RUnlock()
		all := make([]ServiceInfo, 0)
		for _, s := range registry {
			all = append(all, *s)
		}
		c.JSON(http.StatusOK, all)
	})

	r.POST("/admin/services/:name/toggle", func(c *gin.Context) {
		name := c.Param("name")
		var body struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		mu.Lock()
		defer mu.Unlock()
		if s, ok := registry[name]; ok {
			if s.IsCore {
				c.JSON(http.StatusForbidden, gin.H{"error": "cannot disable core service"})
				return
			}
			s.Enabled = body.Enabled
			c.JSON(http.StatusOK, gin.H{"status": "updated", "enabled": s.Enabled})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "not registered"})
	})

	slog.Info("Registry starting", "port", 8088)
	r.Run(":8088")
}
