package main

import (
	"log/slog"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/pkg/tasklib"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("network.db"), &gorm.Config{})
	if err != nil { panic(err) }
	db.AutoMigrate(&models.Subnet{}, &models.IPAddress{})
}

func main() {
	logging.Init("network")
	initDB()
	tasklib.Init("http://localhost:8085")
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

	// Register dummy recurring task for integration test
	tasklib.RegisterEndpoint(
		"http://localhost:8088",
		r,
		"/internal/tasks/dummy-network",
		"@every 10s",
		"http://localhost:8084/internal/tasks/dummy-network",
		"system",
		"dummy_test_network",
		"network-service",
		"IntegrationTest",
		func(payload []byte) error {
			slog.Info("DUMMY_RECURRING_TEST_EXECUTED", "service", "network")
			return nil
		},
	)

	slog.Info("Network starting", "port", 8084)
	r.Run(":8084")
}
