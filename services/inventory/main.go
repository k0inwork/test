// Package main starts the inventory microservice, configuring its Gin router,
// database, background sync engines, and GraphQL endpoints.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/pkg/tasklib"
	"pum-go/pkg/tracing"
	"pum-go/services/inventory/sync"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("inventory.db"), &gorm.Config{})
	if err == nil {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			slog.Error("failed to install gorm otel plugin", "err", err)
		}
	}
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.Switch{}, &models.SwitchPort{}, &models.Ipmi{}, &models.PDU{})
}

func setupRouter(database *gorm.DB, engine *sync.SyncEngine) *gin.Engine {
	r := gin.Default()
	r.Use(otelgin.Middleware("inventory"))
	r.Use(logging.GinMiddleware())

	// Register recurring sync task
	tasklib.RegisterEndpoint(
		"http://localhost:8088", // registry URL
		r,
		"/inventory/task/sync",                       // local webhook path
		"@every 5m",                                  // schedule
		"http://localhost:8083/inventory/task/sync", // target URL reachable by task service
		"system",                                     // username
		"sync-switches",                              // operation
		"inventory-all",                              // object ID
		"Switch",                                     // class name
		func(ctx context.Context, payload []byte) error {
			slog.Info("Executing recurring inventory sync")
			return engine.Run(ctx)
		},
	)

	r.GET("/switches", func(c *gin.Context) {
		var switches []models.Switch
		database.Find(&switches)
		c.JSON(http.StatusOK, switches)
	})

	r.GET("/pdus", func(c *gin.Context) {
		var pdus []models.PDU
		database.Find(&pdus)
		c.JSON(http.StatusOK, pdus)
	})

	r.GET("/ipmi", func(c *gin.Context) {
		var ipmi []models.Ipmi
		database.Find(&ipmi)
		c.JSON(http.StatusOK, ipmi)
	})

	r.POST("/sync", func(c *gin.Context) {
		if err := engine.Run(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	r.POST("/configurable", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	return r
}

func main() {
	tp, _ := tracing.InitTracer("inventory")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "err", err)
		}
	}()
	logging.Init("inventory")
	initDB()
	engine := sync.NewSyncEngine(db, nil)

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "inventory",
		Endpoint: "http://localhost:8083",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "inventory", Endpoints: []string{"/switches", "/pdus", "/ipmi"}},
			{Name: "switches", Endpoints: []string{"/switches"}},
		},
		IsCore:  false,
		OrderID: 3,
		Menu: []logging.MenuItem{
			{Label: "Switches", Path: "/switches"},
		},
	})

	// Initialize tasklib to communicate with the central task microservice
	tasklib.Init("http://localhost:8085")

	r := setupRouter(db, engine)

	slog.Info("Inventory starting", "port", 8083)
	r.Run(":8083")
}
