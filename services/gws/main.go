// Package main represents the deprecated gws (Gateways) microservice entry point.
// Functionality has mostly been migrated, but this remains for backward compatibility.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"pum-go/pkg/config"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/pkg/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

var db *gorm.DB
var GlobalConfig *config.Config

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("gws.db"), &gorm.Config{})
	if err == nil {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			slog.Error("failed to install gorm otel plugin", "err", err)
		}
	}
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.Gw{})
}

func main() {
	tp, _ := tracing.InitTracer("gws")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "err", err)
		}
	}()
	logging.Init("gws")
	cfg, err := config.LoadConfig("system.yaml")
	if err == nil {
		GlobalConfig = cfg
	}

	initDB()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "gws",
		Endpoint: "http://localhost:8091",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "gws", Endpoints: []string{"/"}},
			{Name: "tunnels", Endpoints: []string{"/tunnels"}},
			{Name: "vxlan", Endpoints: []string{"/vxlan"}},
		},
		IsCore:  false,
		OrderID: 6,
		Menu: []logging.MenuItem{
			{Label: "Gateways", Path: "/gateways"},
		},
	})

	r := gin.Default()
	r.Use(otelgin.Middleware("gws"))
	r.Use(logging.GinMiddleware())

	r.GET("/gateways", func(c *gin.Context) {
		var gateways []models.Gw
		db.WithContext(c.Request.Context()).Find(&gateways)
		c.JSON(http.StatusOK, gateways)
	})

	r.POST("/gateways", func(c *gin.Context) {
		var gateway models.Gw
		if err := c.ShouldBindJSON(&gateway); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.WithContext(c.Request.Context()).Create(&gateway)
		c.JSON(http.StatusCreated, gateway)
	})

	slog.Info("GWS service starting", "port", 8091)
	r.Run(":8091")
}
