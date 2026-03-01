package main

import (
	"encoding/json"
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
		// Mimic User list from identity service
		comp.GET("/users", func(c *gin.Context) {
			proxyLegacy(c, "http://localhost:8081/users", "users")
		})

		// Mimic Product/Node list from product service
		comp.GET("/nodes", func(c *gin.Context) {
			proxyLegacy(c, "http://localhost:8082/nodes", "products")
		})
	}

	slog.Info("Compatibility Service starting", "port", 8090)
	r.Run(":8090")
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
		// Mimic the Django json_middleware structure
		c.JSON(http.StatusOK, LegacyResponse{
			Context: map[string]interface{}{
				contextKey: data,
			},
			Errors: []string{},
		})
	} else {
		// For side-by-side comparison, return raw data if json=true is missing
		c.JSON(http.StatusOK, data)
	}
}
