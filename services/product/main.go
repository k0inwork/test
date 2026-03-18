package main

import (
	"context"
	"log/slog"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/pkg/tasklib"
	"pum-go/pkg/tracing"
	"pum-go/services/product/sync"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("product.db"), &gorm.Config{})
	if err == nil {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			slog.Error("failed to install gorm otel plugin", "err", err)
		}
	}
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.Product{}, &models.Gw{}, &models.Session{})
}

func setupRouter(dbConn *gorm.DB, engine *sync.SyncEngine) *gin.Engine {
	r := gin.Default()
	r.Use(otelgin.Middleware("product"))
	r.Use(logging.GinMiddleware())
	db = dbConn

	r.GET("/nodes", func(c *gin.Context) {
		var products []models.Product
		db.WithContext(c.Request.Context()).Find(&products)
		c.JSON(200, products)
	})

	r.GET("/gateways", func(c *gin.Context) {
		var gateways []models.Gw
		db.WithContext(c.Request.Context()).Find(&gateways)
		c.JSON(200, gateways)
	})

	r.POST("/gateways", func(c *gin.Context) {
		var gateway models.Gw
		if err := c.ShouldBindJSON(&gateway); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		db.WithContext(c.Request.Context()).Create(&gateway)
		c.JSON(201, gateway)
	})

	r.GET("/sessions", func(c *gin.Context) {
		var sessions []models.Session
		db.WithContext(c.Request.Context()).Find(&sessions)
		c.JSON(200, sessions)
	})

	r.POST("/sync", func(c *gin.Context) {
		if engine != nil {
			engine.Run(c.Request.Context())
			c.JSON(200, gin.H{"message": "Sync completed"})
		} else {
			c.JSON(400, gin.H{"error": "Sync engine not available"})
		}
	})

	return r
}

func main() {
	tp, _ := tracing.InitTracer("product")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "err", err)
		}
	}()
	logging.Init("product")
	initDB()
	engine := sync.NewSyncEngine(db, nil)

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "product",
		Endpoint: "http://localhost:8082",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "nodes", Endpoints: []string{"/nodes"}},
			{Name: "gateways", Endpoints: []string{"/gateways"}},
			{Name: "sessions", Endpoints: []string{"/sessions"}},
			{Name: "sync", Endpoints: []string{"/sync"}},
		},
		IsCore:  true,
		OrderID: 1,
		Menu: []logging.MenuItem{
			{Label: "Nodes", Path: "/nodes"},
			{Label: "Gateways", Path: "/gateways"},
			{Label: "Sessions", Path: "/sessions"},
		},
	})

	// Initialize tasklib to communicate with the central task microservice
	tasklib.Init("http://localhost:8085")

	r := setupRouter(db, engine)

	// Register recurring sync task
	tasklib.RegisterEndpoint(
		"http://localhost:8088", // registry URL
		r,
		"/product/task/sync",                      // local webhook path
		"@every 1m",                               // schedule
		"http://localhost:8082/product/task/sync", // target URL reachable by task service
		"system",                                  // username
		"sync-products",                           // operation
		"product-all",                             // object ID
		"Product",                                 // class name
		func(ctx context.Context, payload []byte) error {
			slog.Info("Executing recurring product sync")
			return engine.Run(ctx)
		},
	)

	slog.Info("Product starting", "port", 8082)
	r.Run(":8082")
}
