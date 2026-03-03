package main

import (
	"log/slog"
	"net/http"
	"pum-go/pkg/config"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var GlobalConfig *config.Config

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("keyservice.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.KeyService{})
}

func main() {
	logging.Init("keyservice")
	cfg, err := config.LoadConfig("system.yaml")
	if err == nil {
		GlobalConfig = cfg
	}

	initDB()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "keyservice",
		Endpoint:     "http://localhost:8092",
		Capabilities: []string{"services", "key_services", "connectivity"},
		IsCore:       false,
		OrderID:      7,
		Menu: []logging.MenuItem{
			{Label: "Key Services", Path: "/keyservices"},
		},
	})

	r := gin.Default()
	r.Use(logging.GinMiddleware())

	r.GET("/keyservices", func(c *gin.Context) {
		var services []models.KeyService
		db.Find(&services)
		c.JSON(http.StatusOK, services)
	})

	r.POST("/keyservices", func(c *gin.Context) {
		var service models.KeyService
		if err := c.ShouldBindJSON(&service); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		service.Status = "INITIATED"
		service.ConnectionInformation = "1:1"
		db.Create(&service)
		c.JSON(http.StatusCreated, service)
	})

	slog.Info("KeyService starting", "port", 8092)
	r.Run(":8092")
}
