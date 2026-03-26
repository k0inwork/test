// Package main runs the terminal microservice, potentially handling web-based
// SSH sessions or interacting directly with device consoles.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	tp, _ := tracing.InitTracer("terminal")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "err", err)
		}
	}()
	logging.Init("terminal")

	// Register with Registry
	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "terminal",
		Endpoint: "http://localhost:8087",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "terminal", Endpoints: []string{"/"}},
			{Name: "ssh", Endpoints: []string{"/ssh"}},
		},
		IsCore: false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(otelgin.Middleware("terminal"))
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.GET("/session", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Terminal session mocked"})
	})

	slog.Info("Terminal service starting", "port", 8087)
	r.Run(":8087")
}
