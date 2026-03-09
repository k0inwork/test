package main

import (
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
	"pum-go/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"context"

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
	if err == nil {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			slog.Error("failed to install gorm otel plugin", "err", err)
		}
	}
	if err != nil { panic(err) }
	db.AutoMigrate(&models.Product{})
}

func main() {
	tp, _ := tracing.InitTracer("product")
	defer func() { if err := tp.Shutdown(context.Background()); err != nil { slog.Error("failed to shutdown tracer", "err", err) } }()
	logging.Init("product")
	initDB()
	_ = sync.NewSyncEngine(db, nil)

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "product",
		Endpoint:     "http://localhost:8082",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "nodes", Endpoints: []string{"/nodes"}},
			{Name: "sync", Endpoints: []string{"/sync"}},
		},
		IsCore:       true,
		OrderID:      1,
		Menu:         []logging.MenuItem{{Label: "Nodes", Path: "/nodes"}},
	})

	r := gin.Default()
	r.Use(otelgin.Middleware("product"))
	r.Use(logging.GinMiddleware())

	r.GET("/nodes", func(c *gin.Context) {
		var products []models.Product
		db.Find(&products)
		c.JSON(200, products)
	})

	slog.Info("Product starting", "port", 8082)
	r.Run(":8082")
}
