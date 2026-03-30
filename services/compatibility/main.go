// Package main provides the entry point for the compatibility microservice,
// which acts as a bridge for legacy endpoints and legacy JSON mock responses.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var otelClient = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

// LegacyResponse mimics the Django context structure returned by json_middleware
type LegacyResponse struct {
	Context  map[string]interface{} `json:"context"`
	Errors   []string               `json:"errors"`
	Redirect string                 `json:"url_redirect,omitempty"`
}

func main() {
	tp, _ := tracing.InitTracer("compatibility")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "err", err)
		}
	}()
	logging.Init("compatibility")

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "compatibility",
		Endpoint: "http://localhost:8090",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "compatibility", Endpoints: []string{"/"}},
			{Name: "legacy-api", Endpoints: []string{"/legacy-api"}},
		},
		IsCore:  false,
		OrderID: 99,
	})

	r := gin.Default()
	r.Use(otelgin.Middleware("compatibility"))
	r.Use(logging.GinMiddleware())

	// Compatibility Group
	comp := r.Group("/compatibility")
	{
		// --- Ported Services (Active Proxies) ---

		// Identity / Accounts
		comp.GET("/accounts/list/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8081/users", "users") })
		comp.GET("/accounts/currentuser", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8081/users/current", "user") })

		// Products
		comp.GET("/products/products/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8082/nodes", "products") })
		comp.GET("/products/products/:pk/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8082/nodes/"+c.Param("pk"), "product") })

		// --- Legacy Service Stubs (Exact paths from original urls.py) ---

		// products app
		comp.GET("/products/nodes-problems/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8089/problems", "nodes_problems") })
		comp.GET("/products/zabbix_problems/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8089/problems", "zabbix_problems") })

		// data app
		comp.GET("/data/switch/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8083/switches", "switch_list") })
		comp.GET("/data/pdu/list", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8083/pdus", "pdu_list") })
		comp.GET("/data/ipmi/list", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8083/ipmi", "ipmi_list") })

		// network app
		comp.GET("/network/dhcp/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8084/dhcp", "dhcp_list") })
		comp.GET("/network/dns/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8084/dns", "dns_list") })
		comp.GET("/network/subnetwork/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8084/subnets", "subnetwork_list") })


		// services app
		comp.GET("/services/listkeyservice/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8092/keyservices", "key_services") })
		comp.GET("/services/listdataservice/", stubHandler("data_services"))

		// tasks app
		comp.GET("/tasks/viewtasks/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8085/tasks", "tasks") })

		// accounts app (extra)
		comp.GET("/accounts/activitylist/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8081/activitylist", "activity_list") })
		comp.GET("/accounts/group/list", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8081/groups", "groups") })
	}

	slog.Info("Compatibility Service starting", "port", 8090)
	r.Run(":8090")
}

func stubHandler(contextKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		isJSON := c.Query("json") == "true"
		if isJSON {
			c.JSON(http.StatusOK, LegacyResponse{
				Context: map[string]interface{}{
					contextKey: []interface{}{},
					"message":  fmt.Sprintf("Legacy endpoint for %s is not yet implemented in Go microservices.", contextKey),
				},
				Errors: []string{},
			})
		} else {
			c.String(http.StatusNotImplemented, "Endpoint %s not yet implemented", contextKey)
		}
	}
}

func proxyLegacy(c *gin.Context, targetURL string, contextKey string) {
	isJSON := c.Query("json") == "true"

	req, _ := http.NewRequestWithContext(c.Request.Context(), "GET", targetURL, nil)
	resp, err := otelClient.Do(req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Target service unreachable"})
		return
	}
	defer resp.Body.Close()

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode target response"})
		return
	}

	if isJSON {
		c.JSON(http.StatusOK, LegacyResponse{
			Context: map[string]interface{}{
				contextKey: data,
			},
			Errors: []string{},
		})
	} else {
		c.JSON(http.StatusOK, data)
	}
}
