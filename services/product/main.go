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

func setupRouter(dbConn *gorm.DB, engine *sync.SyncEngine) *gin.Engine {
	r := gin.Default()
	r.Use(logging.GinMiddleware())
	db = dbConn

	r.GET("/nodes", func(c *gin.Context) {
		var products []models.Product
		db.Find(&products)
		c.JSON(200, products)
	})

	r.GET("/gateways", func(c *gin.Context) {
		var gateways []models.Product
		// In the new architecture, gateways are just nodes with pou_type='GW'
		// or we can just return all nodes if the GUI expects that.
		// For backward compatibility with the GWS service, we'll try to filter.
		db.Where("pou_type = ?", "GW").Find(&gateways)
		c.JSON(200, gateways)
	})

	r.POST("/gateways", func(c *gin.Context) {
		var gateway models.Product
		if err := c.ShouldBindJSON(&gateway); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		gateway.PouType = "GW"
		db.Create(&gateway)
		c.JSON(201, gateway)
	})

	r.POST("/sync", func(c *gin.Context) {
		engine.Run()
		c.JSON(200, gin.H{"message": "Sync completed"})
	})

	return r
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
			{Name: "gateways", Endpoints: []string{"/gateways"}},
			{Name: "sync", Endpoints: []string{"/sync"}},
		},
		IsCore:  true,
		OrderID: 1,
		Menu: []logging.MenuItem{
			{Label: "Nodes", Path: "/nodes"},
			{Label: "Gateways", Path: "/gateways"},
		},
	})

	// Initialize tasklib to communicate with the central task microservice
	tasklib.Init("http://localhost:8085")

	r := setupRouter(db, engine)

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

	slog.Info("Product starting", "port", 8082)
	r.Run(":8082")
}
