package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"pum-go/pkg/config"
	"pum-go/pkg/logging"
	"pum-go/pkg/tracing"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type ModuleRequest struct {
	TargetIP string `json:"target_ip"`
	Command  string `json:"command"`
	Param    string `json:"param"`
}

func main() {
	tp, _ := tracing.InitTracer("external-modules")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "err", err)
		}
	}()
	logging.Init("external-modules")

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "external-modules",
		Endpoint: "http://localhost:8086",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "external-modules", Endpoints: []string{"/"}},
			{Name: "pdu", Endpoints: []string{"/pdu"}},
			{Name: "ipmi", Endpoints: []string{"/ipmi"}},
			{Name: "network-management", Endpoints: []string{"/network-management"}},
			{Name: "routing", Endpoints: []string{"/routing"}},
			{Name: "configurable", Endpoints: []string{"/configurable"}},
		},
		IsCore: false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(otelgin.Middleware("external-modules"))
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	// Configurable endpoint to receive system configuration
	r.POST("/configurable", func(c *gin.Context) {
		var cfg config.Config
		if err := c.ShouldBindJSON(&cfg); err != nil {
			slog.Error("Failed to parse configuration push", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid configuration payload"})
			return
		}

		slog.Info("Successfully received configuration from registry", "external_modules_count", len(cfg.ExternalModules))
		// Apply configuration here if needed, for example updating endpoints or modes
		c.JSON(http.StatusOK, gin.H{"status": "configuration applied"})
	})

	// Generic Module Call (RabbitMQ Proxy Placeholder)
	r.POST("/call", func(c *gin.Context) {
		var req ModuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		slog.Info("Proxying request to external module", "ip", req.TargetIP, "cmd", req.Command)

		// Simulate RabbitMQ/Remote call delay
		time.Sleep(200 * time.Millisecond)

		c.JSON(http.StatusOK, gin.H{
			"status":   "success",
			"response": fmt.Sprintf("Module command %s executed on %s", req.Command, req.TargetIP),
		})
	})

	// Specific PDU Control
	r.POST("/pdu/relay", func(c *gin.Context) {
		var req ModuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		slog.Info("Executing PDU command via module", "ip", req.TargetIP, "cmd", req.Command, "outlet", req.Param)
		time.Sleep(500 * time.Millisecond)

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": fmt.Sprintf("PDU Command %s executed on %s (Outlet %s)", req.Command, req.TargetIP, req.Param),
		})
	})

	// Specific IPMI Control
	r.POST("/ipmi/power", func(c *gin.Context) {
		var req ModuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		slog.Info("Executing IPMI power command via module", "ip", req.TargetIP, "cmd", req.Command)
		time.Sleep(1 * time.Second)

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": fmt.Sprintf("IPMI Command %s initiated on %s", req.Command, req.TargetIP),
		})
	})

	slog.Info("External Modules Proxy starting", "port", 8086)
	r.Run(":8086")
}
