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

type LegacyResponse struct {
	Context  map[string]interface{} `json:"context"`
	Errors   []string               `json:"errors"`
	Redirect string                 `json:"url_redirect,omitempty"`
}

func main() {
	tp, _ := tracing.InitTracer("compatibility")
	defer func() { tp.Shutdown(context.Background()) }()
	logging.Init("compatibility")

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "compatibility",
		Endpoint: "http://localhost:8090",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "compatibility", Endpoints: []string{"/"}},
		},
		IsCore:  false,
		OrderID: 99,
	})

	r := gin.Default()
	r.Use(otelgin.Middleware("compatibility"))
	r.Use(logging.GinMiddleware())

	comp := r.Group("/compatibility")
	{
		comp.GET("/accounts/list/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8081/users", "users") })
		comp.GET("/accounts/currentuser", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8081/users/current", "user") })
		comp.GET("/products/products/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8082/nodes", "products") })
		comp.GET("/products/products/:pk/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8082/nodes/"+c.Param("pk"), "product") })

		// GWS logic now served by Product microservice (port 8082)
		comp.GET("/gws/gws/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8082/gateways", "gateways") })
		comp.GET("/gws/historysession/", func(c *gin.Context) { proxyLegacy(c, "http://localhost:8082/sessions", "session_history") })

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
	json.NewDecoder(resp.Body).Decode(&data)

	if isJSON {
		c.JSON(http.StatusOK, LegacyResponse{
			Context: map[string]interface{}{contextKey: data},
			Errors:  []string{},
		})
	} else {
		c.JSON(http.StatusOK, data)
	}
}
