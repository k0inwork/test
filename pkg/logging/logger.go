// Package logging offers standardized, structured logging and HTTP middleware
// for auditing and request tracking across all microservices using the Go standard library.
package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

import (
	"sync"
)

var otelClient = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

var (
	L                *slog.Logger
	registryEndpoint string
	auditEndpoint    string
	auditMu          sync.RWMutex
)

type MenuItem struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

type CapabilityRegistration struct {
	Name      string   `json:"name"`
	Endpoints []string `json:"endpoints"`
}

type ServiceRegistration struct {
	Name         string                   `json:"name"`
	Endpoint     string                   `json:"endpoint"`
	Capabilities []CapabilityRegistration `json:"capabilities"`
	IsCore       bool                     `json:"is_core"`
	OrderID      int                      `json:"order_id"`
	Menu         []MenuItem               `json:"menu"`
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
	registryEndpoint = registryURL
	go func() {
		backoff := 1 * time.Second
		maxBackoff := 30 * time.Second

		for {
			data, _ := json.Marshal(info)
			resp, err := otelClient.Post(registryURL+"/register", "application/json", bytes.NewBuffer(data))
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				slog.Info("Registered with service discovery", "url", registryURL)
				break
			}
			if err == nil {
				resp.Body.Close()
			}
			slog.Warn("Failed to register with discovery, retrying...", "error", err)
			time.Sleep(backoff)

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
		for {
			time.Sleep(30 * time.Second)
			resp, err := otelClient.Post(registryURL+"/heartbeat/"+info.Name, "application/json", nil)
			if err == nil {
				resp.Body.Close()
			}
		}
	}()
}

func getAuditEndpoint() string {
	auditMu.RLock()
	ep := auditEndpoint
	auditMu.RUnlock()
	if ep != "" {
		return ep
	}

	if registryEndpoint == "" {
		return ""
	}

	resp, err := otelClient.Get(registryEndpoint + "/capabilities/audit")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var res struct {
			Endpoint string `json:"endpoint"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err == nil && res.Endpoint != "" {
			auditMu.Lock()
			auditEndpoint = res.Endpoint
			auditMu.Unlock()
			return res.Endpoint
		}
	}
	return ""
}

func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		queryParams := c.Request.URL.RawQuery

		c.Next()
		latency := time.Since(start)

		status := c.Writer.Status()

		L.InfoContext(c.Request.Context(), "request handled",
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.Duration("latency", latency),
		)

		// Send audit log asynchronously
		go func() {
			ep := getAuditEndpoint()
			if ep == "" {
				return
			}

			// Don't log requests to the audit endpoint itself to avoid loops
			if path == "/activitylist" {
				return
			}

			username, err := c.Cookie("pum_user")
			if err != nil || username == "" {
				username = "system"
			}

			logEntry := map[string]interface{}{
				"username":       username,
				"request_method": method,
				"request_url":    path,
				"query_params":   queryParams,
				"response_code":  status,
			}

			data, _ := json.Marshal(logEntry)
			otelClient.Post(ep, "application/json", bytes.NewBuffer(data))
		}()
	}
}

func WaitForService(registryURL, targetServiceName string) {
	for {
		resp, err := otelClient.Get(registryURL + "/services")
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
