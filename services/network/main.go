package main

import (
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"

	"github.com/gin-gonic/gin"
)

func main() {
	logging.Init("network")

	// Register with Registry
	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "network",
		Endpoint:     "http://localhost:8084",
		Capabilities: []string{"network", "routing"},
		IsCore:       false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.GET("/routes", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Network routing logic (Mocked)"})
	})

	slog.Info("Network service starting", "port", 8084)
	r.Run(":8084")
}
