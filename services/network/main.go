package main

import (
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
	"pum-go/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"context"

	"log/slog"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("network.db"), &gorm.Config{})
	if err == nil {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			slog.Error("failed to install gorm otel plugin", "err", err)
		}
	}
	if err != nil { panic(err) }
	db.AutoMigrate(&models.Subnet{}, &models.IPAddress{})
}

func main() {
	tp, _ := tracing.InitTracer("network")
	defer func() { if err := tp.Shutdown(context.Background()); err != nil { slog.Error("failed to shutdown tracer", "err", err) } }()
	logging.Init("network")
	initDB()
	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "network",
		Endpoint:     "http://localhost:8084",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "network", Endpoints: []string{"/"}},
			{Name: "ipam", Endpoints: []string{"/ipam"}},
			{Name: "routing", Endpoints: []string{"/routing"}},
		},
		IsCore:       false,
		OrderID:      3,
		Menu:         []logging.MenuItem{{Label: "Subnets", Path: "/subnets"}},
	})
	r := gin.Default()
	r.Use(otelgin.Middleware("network"))
	r.Use(logging.GinMiddleware())
	r.GET("/subnets", func(c *gin.Context) {
		var subnets []models.Subnet
		db.Find(&subnets)
		c.JSON(200, subnets)
	})

	// DHCP Mock list
	r.GET("/dhcp", func(c *gin.Context) {
		// Mock response simulating RabbitMQ "dhcp host list" RPC result
		c.JSON(200, []map[string]interface{}{
			{"ip": "10.10.1.50", "mac": "00:1A:2B:3C:4D:5E", "hostname": "client-pc-1"},
		})
	})

	// DNS Mock list
	r.GET("/dns", func(c *gin.Context) {
		// Mock response simulating RabbitMQ "dns host list" RPC result
		c.JSON(200, []map[string]interface{}{
			{"name": "server.local", "ip": "10.10.1.100", "type": "A"},
			{"name": "router.local", "ip": "10.10.1.1", "type": "A"},
		})
	})

	slog.Info("Network starting", "port", 8084)
	r.Run(":8084")
}
