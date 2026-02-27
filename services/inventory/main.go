package main

import (
	"fmt"
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

	var count int64
	db.Model(&models.Switch{}).Count(&count)
	if count == 0 {
		for i := 1; i <= 10; i++ {
			swID := fmt.Sprintf("sw-%d", i)
			sw := models.Switch{
				ID: swID, Name: fmt.Sprintf("Switch-%02d", i),
				IP: fmt.Sprintf("10.10.0.%d", i),
				Model: "Cisco Catalyst 9300",
				LogicalType: "cl", PortsCount: 48,
			}
			db.Create(&sw)

			// Create 3 ports for each switch
			for p := 1; p <= 3; p++ {
				db.Create(&models.SwitchPort{
					ID: fmt.Sprintf("p-%d-%d", i, p),
					SwitchID: swID,
					Port: fmt.Sprintf("GigabitEthernet1/0/%d", p),
					Vlan: 10 * p,
				})
			}
		}
	}
}

func main() {
	logging.Init("inventory")
	initDB()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "inventory",
		Endpoint:     "http://localhost:8083",
		Capabilities: []string{"inventory", "switches", "ports"},
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

	r.GET("/switches/:id", func(c *gin.Context) {
		var sw models.Switch
		if err := db.First(&sw, "id = ?", c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Switch not found"})
			return
		}
		c.JSON(http.StatusOK, sw)
	})

	r.GET("/switches/:id/ports", func(c *gin.Context) {
		var ports []models.SwitchPort
		db.Where("switch_id = ?", c.Param("id")).Find(&ports)
		c.JSON(http.StatusOK, ports)
	})

	r.POST("/switches", func(c *gin.Context) {
		var sw models.Switch
		if err := c.ShouldBindJSON(&sw); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Create(&sw)
		c.JSON(http.StatusCreated, sw)
	})

	r.PUT("/ports/:id/vlan", func(c *gin.Context) {
		var input struct {
			Vlan int `json:"vlan"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db.Model(&models.SwitchPort{}).Where("id = ?", c.Param("id")).Update("vlan", input.Vlan)
		slog.Info("VLAN updated on port", "id", c.Param("id"), "vlan", input.Vlan)
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	})

	slog.Info("Inventory service starting", "port", 8083)
	r.Run(":8083")
}
