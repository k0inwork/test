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
	db, err = gorm.Open(sqlite.Open("inventory.db"), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		panic(err)
	}

	db.AutoMigrate(&models.Switch{}, &models.SwitchPort{})
}

func main() {
	logging.Init("inventory")
	initDB()

	// Register with Registry
	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "inventory",
		Endpoint:     "http://localhost:8083",
		Capabilities: []string{"inventory", "switches"},
		IsCore:       false,
	})

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logging.GinMiddleware())

	r.GET("/switches", func(c *gin.Context) {
		var switches []models.Switch
		db.Find(&switches)
		c.JSON(http.StatusOK, switches)
	})

	slog.Info("Inventory service starting", "port", 8083)
	r.Run(":8083")
}
