package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var L *slog.Logger

type ServiceRegistration struct {
	Name         string   `json:"name"`
	Endpoint     string   `json:"endpoint"`
	Capabilities []string `json:"capabilities"`
	IsCore       bool     `json:"is_core"`
}

func Init(serviceName string) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs([]slog.Attr{
		slog.String("service", serviceName),
	})

	L = slog.New(handler)
	slog.SetDefault(L)
}

// RegisterWithDiscovery informs the Registry service about the current service
func RegisterWithDiscovery(registryURL string, info ServiceRegistration) {
	go func() {
		for {
			data, _ := json.Marshal(info)
			resp, err := http.Post(registryURL+"/register", "application/json", bytes.NewBuffer(data))
			if err == nil {
				resp.Body.Close()
				slog.Info("Registered with service discovery", "url", registryURL)
				break
			}
			slog.Warn("Failed to register with discovery, retrying...", "error", err)
			time.Sleep(5 * time.Second)
		}

		// Keep alive heartbeat
		for {
			time.Sleep(30 * time.Second)
			resp, err := http.Post(registryURL+"/heartbeat/"+info.Name, "application/json", nil)
			if err != nil {
				slog.Warn("Heartbeat failed", "service", info.Name, "error", err)
			} else {
				resp.Body.Close()
			}
		}
	}()
}

// GinMiddleware returns a gin.HandlerFunc that logs requests
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		L.InfoContext(c.Request.Context(), "request handled",
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.String("user_agent", c.Request.UserAgent()),
		)
	}
}

// GetLogger returns the global logger
func GetLogger() *slog.Logger {
	if L == nil {
		Init("default")
	}
	return L
}
