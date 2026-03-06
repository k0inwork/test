package main

import (
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
	"pum-go/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"context"

	"log/slog"
	"net/http"
	"pum-go/pkg/config"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var GlobalConfig *config.Config

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("keyservice.db"), &gorm.Config{})
	if err == nil {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			slog.Error("failed to install gorm otel plugin", "err", err)
		}
	}
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.KeyService{})
}

func main() {
	tp, _ := tracing.InitTracer("keyservice")
	defer func() { if err := tp.Shutdown(context.Background()); err != nil { slog.Error("failed to shutdown tracer", "err", err) } }()
	logging.Init("keyservice")
	cfg, err := config.LoadConfig("system.yaml")
	if err == nil {
		GlobalConfig = cfg
	}

	initDB()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "keyservice",
		Endpoint:     "http://localhost:8092",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "services", Endpoints: []string{"/"}},
			{Name: "key_services", Endpoints: []string{"/key_services"}},
			{Name: "connectivity", Endpoints: []string{"/connectivity"}},
		},
		IsCore:       false,
		OrderID:      7,
		Menu: []logging.MenuItem{
			{Label: "Key Services", Path: "/keyservices"},
		},
	})

	r := gin.Default()
	r.Use(otelgin.Middleware("keyservice"))
	r.Use(logging.GinMiddleware())

	r.GET("/keyservices", func(c *gin.Context) {
		var services []models.KeyService
		db.Find(&services)
		c.JSON(http.StatusOK, services)
	})

	r.POST("/keyservices", func(c *gin.Context) {
		var service models.KeyService
		if err := c.ShouldBindJSON(&service); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		service.Status = "INITIATED"
		service.ConnectionInformation = "1:1"
		db.Create(&service)
		c.JSON(http.StatusCreated, service)
	})

	slog.Info("KeyService starting", "port", 8092)
	r.Run(":8092")
}
