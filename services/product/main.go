package main

import (
	"log/slog"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/pkg/tasklib"
	"pum-go/services/product/sync"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("product.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.Product{})
}

func main() {
	logging.Init("product")
	initDB()
	engine := sync.NewSyncEngine(db, nil)

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "product",
		Endpoint:     "http://localhost:8082",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "nodes", Endpoints: []string{"/nodes"}},
			{Name: "sync", Endpoints: []string{"/sync"}},
		},
		IsCore:  true,
		OrderID: 1,
		Menu:    []logging.MenuItem{{Label: "Nodes", Path: "/nodes"}},
	})

	// Initialize tasklib to communicate with the central task microservice
	tasklib.Init("http://localhost:8085")

	r := gin.Default()
	r.Use(logging.GinMiddleware())

	// Register recurring sync task
	tasklib.RegisterEndpoint(
		"http://localhost:8088", // registry URL
		r,
		"/product/task/sync", // local webhook path
		"@every 1m",            // schedule
		"http://localhost:8082/product/task/sync", // target URL reachable by task service
		"system",               // username
		"sync-products",        // operation
		"product-all",          // object ID
		"Product",              // class name
		func(payload []byte) error {
			slog.Info("Executing recurring product sync")
			return engine.Run()
		},
	)

	r.GET("/nodes", func(c *gin.Context) {
		var products []models.Product
		db.Find(&products)
		c.JSON(200, products)
	})

	slog.Info("Product starting", "port", 8082)
	r.Run(":8082")
}
