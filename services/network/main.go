package main

import (
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
	if err != nil { panic(err) }
	db.AutoMigrate(&models.Subnet{}, &models.IPAddress{})
}

func main() {
	logging.Init("network")
	initDB()
	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "network",
		Endpoint:     "http://localhost:8084",
		Capabilities: []string{"network", "ipam", "routing"},
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
	slog.Info("Network starting", "port", 8084)
	r.Run(":8084")
}
