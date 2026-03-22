package main

import (
	"context"
	"log/slog"
	"net/http"
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

func initDB(dialector gorm.Dialector) *gorm.DB {
	d, err := gorm.Open(dialector, &gorm.Config{})
	if err == nil {
		if err := d.Use(otelgorm.NewPlugin()); err != nil {
			slog.Error("failed to install gorm otel plugin", "err", err)
		}
	}
	if err != nil {
		panic(err)
	}
	d.AutoMigrate(&models.Product{}, &models.Gw{}, &models.Session{})
	return d
}

func setupRouter(database *gorm.DB, engine *sync.SyncEngine) *gin.Engine {
	r := gin.Default()
	r.Use(otelgin.Middleware("product"))
	r.Use(logging.GinMiddleware())

	// Register recurring sync task
	tasklib.RegisterEndpoint(
		"http://localhost:8088", // registry URL
		r,
		"/product/task/sync",                       // local webhook path
		"@every 5m",                                // schedule
		"http://localhost:8082/product/task/sync", // target URL reachable by task service
		"system",                                   // username
		"sync-products",                            // operation
		"product-all",                              // object ID
		"Product",                                  // class name
		func(ctx context.Context, payload []byte) error {
			slog.Info("Executing recurring product sync")
			return engine.Run(ctx)
		},
	)

	r.GET("/nodes", func(c *gin.Context) {
		var products []models.Product
		database.Find(&products)
		c.JSON(http.StatusOK, products)
	})

	r.GET("/gateways", func(c *gin.Context) {
		var gws []models.Gw
		database.Find(&gws)
		c.JSON(http.StatusOK, gws)
	})

	r.POST("/gateways", func(c *gin.Context) {
		var gw models.Gw
		if err := c.ShouldBindJSON(&gw); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		database.Save(&gw)
		c.JSON(http.StatusOK, gw)
	})

	r.GET("/sessions", func(c *gin.Context) {
		var sessions []models.Session
		database.Find(&sessions)
		c.JSON(http.StatusOK, sessions)
	})

	r.POST("/sync", func(c *gin.Context) {
		if err := engine.Run(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
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
	db = initDB(sqlite.Open("product.db"))
	engine := sync.NewSyncEngine(db, nil)

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "product",
		Endpoint: "http://localhost:8082",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "nodes", Endpoints: []string{"/nodes"}},
			{Name: "gateways", Endpoints: []string{"/gateways"}},
			{Name: "sessions", Endpoints: []string{"/sessions"}},
		},
		Menu: []logging.MenuItem{
			{Label: "Nodes", Path: "/nodes"},
			{Label: "Gateways", Path: "/gateways"},
		},
		IsCore:  true,
		OrderID: 1,
	})

	// Initialize tasklib to communicate with the central task microservice
	tasklib.Init("http://localhost:8085")

	r := setupRouter(db, engine)

	slog.Info("Product starting", "port", 8082)
	r.Run(":8082")
}
