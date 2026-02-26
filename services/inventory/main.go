package main

import (
	"log"
	"net/http"
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
		log.Fatal("failed to connect database")
	}

	db.AutoMigrate(&models.Switch{}, &models.SwitchPort{})
}

func main() {
	initDB()

	r := gin.Default()

	r.GET("/switches", func(c *gin.Context) {
		var switches []models.Switch
		db.Find(&switches)
		c.JSON(http.StatusOK, switches)
	})

	log.Println("Inventory service starting on :8083")
	r.Run(":8083")
}
