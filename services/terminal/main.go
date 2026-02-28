package main

import (
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"

	"github.com/gin-gonic/gin"
)

func main() {
	logging.Init("terminal")

	// Register with Registry
	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "terminal",
		Endpoint:     "http://localhost:8087",
		Capabilities: []string{"terminal", "ssh"},
		IsCore:       false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.GET("/session", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Terminal session mocked"})
	})

	slog.Info("Terminal service starting", "port", 8087)
	r.Run(":8087")
}
