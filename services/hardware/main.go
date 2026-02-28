package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"time"

	"github.com/gin-gonic/gin"
)

type HardwareRequest struct {
	TargetIP string `json:"target_ip"`
	Command  string `json:"command"`
	Param    string `json:"param"`
}

func main() {
	logging.Init("hardware")

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "hardware",
		Endpoint:     "http://localhost:8086",
		Capabilities: []string{"hardware", "pdu", "ipmi"},
		IsCore:       false,
	})

func main() {
	logging.Init("hardware")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	// PDU Control
	r.POST("/pdu/relay", func(c *gin.Context) {
		var req HardwareRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		slog.Info("Executing PDU command", "ip", req.TargetIP, "cmd", req.Command, "outlet", req.Param)

		// Simulate hardware delay
		time.Sleep(500 * time.Millisecond)

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"message": fmt.Sprintf("Command %s executed on %s (Outlet %s)", req.Command, req.TargetIP, req.Param),
		})
	})

	// IPMI Control
	r.POST("/ipmi/power", func(c *gin.Context) {
		var req HardwareRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		slog.Info("Executing IPMI power command", "ip", req.TargetIP, "cmd", req.Command)

		// Simulate IPMI boot delay
		time.Sleep(1 * time.Second)

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"message": fmt.Sprintf("Server %s: Power %s initiated", req.TargetIP, req.Command),
	r.POST("/call", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"response": "Hardware call mocked",
		})
	})

	slog.Info("Hardware service starting", "port", 8086)
	r.Run(":8086")
}
