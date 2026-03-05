package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var L *slog.Logger

type MenuItem struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

type ServiceRegistration struct {
	Name         string     `json:"name"`
	Endpoint     string     `json:"endpoint"`
	Capabilities []string   `json:"capabilities"`
	IsCore       bool       `json:"is_core"`
	OrderID      int        `json:"order_id"`
	Menu         []MenuItem `json:"menu"`
}

func Init(serviceName string) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	handler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs([]slog.Attr{
		slog.String("service", serviceName),
	})
	L = slog.New(handler)
	slog.SetDefault(L)
}

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
		for {
			time.Sleep(30 * time.Second)
			resp, err := http.Post(registryURL+"/heartbeat/"+info.Name, "application/json", nil)
			if err == nil {
				resp.Body.Close()
			}
		}
	}()
}

// WaitForService blocks the current goroutine until the target service is marked
// as active and enabled in the registry. It checks every 5 seconds.
func WaitForService(registryURL, targetServiceName string) {
	client := &http.Client{Timeout: 5 * time.Second}
	for {
		resp, err := client.Get(registryURL + "/services")
		if err == nil && resp.StatusCode == http.StatusOK {
			var services []ServiceRegistration
			if err := json.NewDecoder(resp.Body).Decode(&services); err == nil {
				for _, s := range services {
					if s.Name == targetServiceName {
						resp.Body.Close()
						slog.Info("Service is now available", "service", targetServiceName)
						return
					}
				}
			}
			resp.Body.Close()
		}
		slog.Info("Waiting for service to become available...", "service", targetServiceName)
		time.Sleep(5 * time.Second)
	}
}

func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		latency := time.Since(start)
		L.InfoContext(c.Request.Context(), "request handled",
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Duration("latency", latency),
		)
	}
}
