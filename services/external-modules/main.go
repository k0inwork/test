package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"time"

	"github.com/gin-gonic/gin"
)

type ModuleRequest struct {
	TargetIP string `json:"target_ip"`
	Command  string `json:"command"`
	Param    string `json:"param"`
}

func main() {
	logging.Init("external-modules")

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "external-modules",
		Endpoint:     "http://localhost:8086",
		Capabilities: []string{"external-modules", "pdu", "ipmi", "network-management", "routing"},
		IsCore:       false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

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
			"status": "success",
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
			"status": "success",
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
			"status": "success",
			"message": fmt.Sprintf("IPMI Command %s initiated on %s", req.Command, req.TargetIP),
		})
	})

	slog.Info("External Modules Proxy starting", "port", 8086)
	r.Run(":8086")
}
