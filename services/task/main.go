package main

import (
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
	"pum-go/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"context"

	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("task.db"), &gorm.Config{})
	if err == nil {
		if err := db.Use(otelgorm.NewPlugin()); err != nil {
			slog.Error("failed to install gorm otel plugin", "err", err)
		}
	}
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.TaskRecord{})
}

func main() {
	tp, _ := tracing.InitTracer("task")
	defer func() { if err := tp.Shutdown(context.Background()); err != nil { slog.Error("failed to shutdown tracer", "err", err) } }()
	logging.Init("task")
	initDB()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:         "task",
		Endpoint:     "http://localhost:8085",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "tasks", Endpoints: []string{"/tasks"}},
			{Name: "async-executor", Endpoints: []string{"/async-executor"}},
		},
		IsCore:       false,
		OrderID:      5,
		Menu: []logging.MenuItem{
			{Label: "Tasks", Path: "/tasks"},
		},
	})

	r := gin.Default()
	r.Use(otelgin.Middleware("task"))
	r.Use(logging.GinMiddleware())

	r.GET("/manifest", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"OrderID": 5,
			"Menu": []logging.MenuItem{
				{Label: "Tasks", Path: "/tasks"},
			},
		})
	})

	r.GET("/tasks", func(c *gin.Context) {
		var tasks []models.TaskRecord
		db.Order("started_at desc").Find(&tasks)
		c.JSON(http.StatusOK, tasks)
	})

	slog.Info("Task service starting", "port", 8085)
	r.Run(":8085")
}
