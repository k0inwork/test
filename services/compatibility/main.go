package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"

	"github.com/gin-gonic/gin"
)

// LegacyResponse mimics the Django context structure returned by json_middleware
type LegacyResponse struct {
	Context map[string]interface{} `json:"context"`
	Errors  []string               `json:"errors"`
	Redirect string                `json:"url_redirect,omitempty"`
}

func main() {
	logging.Init("compatibility")

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "compatibility",
		Endpoint:     "http://localhost:8090",
		Capabilities: []string{"compatibility", "legacy-api"},
		IsCore:       false,
		OrderID:      99,
	})

	r := gin.Default()
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
		comp.GET("/products/nodes-problems/", stubHandler("nodes_problems"))
		comp.GET("/products/zabbix_problems/", stubHandler("zabbix_problems"))

		// data app
		comp.GET("/data/switch/", stubHandler("switch_list"))
		comp.GET("/data/pdu/list", stubHandler("pdu_list"))
		comp.GET("/data/ipmi/list", stubHandler("ipmi_list"))

		// network app
		comp.GET("/network/dhcp/", stubHandler("dhcp_list"))
		comp.GET("/network/dns/", stubHandler("dns_list"))
		comp.GET("/network/subnetwork/", stubHandler("subnetwork_list"))

		// gws app
		comp.GET("/gws/gws/", stubHandler("gateways"))
		comp.GET("/gws/historysession/", stubHandler("session_history"))

		// services app
		comp.GET("/services/listkeyservice/", stubHandler("key_services"))
		comp.GET("/services/listdataservice/", stubHandler("data_services"))

		// tasks app
		comp.GET("/tasks/viewtasks/", stubHandler("tasks"))

		// accounts app (extra)
		comp.GET("/accounts/activitylist/", stubHandler("activity_list"))
		comp.GET("/accounts/group/list", stubHandler("groups"))
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
					"message":    fmt.Sprintf("Legacy endpoint for %s is not yet implemented in Go microservices.", contextKey),
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

	resp, err := http.Get(targetURL)
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
