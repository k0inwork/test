package main

import (
	"log/slog"
	"net/http"
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
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		panic(err)
	}

	db.AutoMigrate(&models.Subnet{}, &models.IPAddress{})

	// Seed
	var count int64
	db.Model(&models.Subnet{}).Count(&count)
	if count == 0 {
		s := models.Subnet{Prefix: "10.0.0.0/24", Region: "MSK", Type: "management"}
		db.Create(&s)
		db.Create(&models.IPAddress{Address: "10.0.0.1", SubnetID: s.ID, Status: "gateway"})
		db.Create(&models.IPAddress{Address: "10.0.0.10", SubnetID: s.ID, Status: "allocated"})
	}
}

func main() {
	logging.Init("network")
	initDB()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "network",
		Endpoint:     "http://localhost:8084",
		Capabilities: []string{"network", "ipam", "routing"},
		IsCore:       false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.GET("/subnets", func(c *gin.Context) {
		var subnets []models.Subnet
		db.Find(&subnets)
		c.JSON(http.StatusOK, subnets)
	})

	r.GET("/subnets/:id/ips", func(c *gin.Context) {
		var ips []models.IPAddress
		db.Where("subnet_id = ?", c.Param("id")).Find(&ips)
		c.JSON(http.StatusOK, ips)
	})

	r.GET("/routes", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Network routing logic (Mocked)"})
	})

	slog.Info("Network service starting", "port", 8084)
	r.Run(":8084")
}
