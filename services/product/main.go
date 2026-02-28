package main

import (
	"log/slog"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/services/product/sync"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("product.db"), &gorm.Config{})
	if err != nil { panic(err) }
	db.AutoMigrate(&models.Product{})
}

func main() {
	logging.Init("product")
	initDB()
	_ = sync.NewSyncEngine(db, nil)

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "product",
		Endpoint:     "http://localhost:8082",
		Capabilities: []string{"nodes", "sync"},
		IsCore:       true,
		OrderID:      1,
		Menu:         []logging.MenuItem{{Label: "Nodes", Path: "/nodes"}},
	})

	r := gin.Default()
	r.Use(logging.GinMiddleware())

	r.GET("/nodes", func(c *gin.Context) {
		var products []models.Product
		db.Find(&products)
		c.JSON(200, products)
	})

	slog.Info("Product starting", "port", 8082)
	r.Run(":8082")
}
