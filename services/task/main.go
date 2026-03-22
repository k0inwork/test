package main

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/pkg/tracing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	otelgorm "gorm.io/plugin/opentelemetry/tracing"
)

var otelClient = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
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
	db.AutoMigrate(&models.TaskRecord{}, &models.Alarm{}, &models.RecurringTask{})
}

func listRecurringTasks(c *gin.Context) {
	var tasks []models.RecurringTask
	db.Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

func createRecurringTask(c *gin.Context) {
	var rt models.RecurringTask
	if err := c.ShouldBindJSON(&rt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.RecurringTask
	if db.Where("name = ?", rt.Name).First(&existing).Error == nil {
		rt.ID = existing.ID
		db.Save(&rt)
		c.JSON(http.StatusOK, rt)
	} else {
		db.Create(&rt)
		c.JSON(http.StatusCreated, rt)
	}
}

func main() {
	tp, _ := tracing.InitTracer("task")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "err", err)
		}
	}()
	logging.Init("task")
	initDB()

	scheduler := cron.New()
	loadRecurringTasks(scheduler)
	scheduler.Start()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "task",
		Endpoint: "http://localhost:8085",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "tasks", Endpoints: []string{"/tasks"}},
			{Name: "recurring-tasks", Endpoints: []string{"/recurring-tasks"}},
		},
		IsCore:  false,
		OrderID: 5,
		Menu: []logging.MenuItem{
			{Label: "Tasks", Path: "/tasks"},
		},
	})

	r := gin.Default()
	r.Use(otelgin.Middleware("task"))
	r.Use(logging.GinMiddleware())

	r.GET("/tasks", func(c *gin.Context) {
		var tasks []models.TaskRecord
		db.Find(&tasks)
		c.JSON(200, tasks)
	})

	r.GET("/recurring-tasks", listRecurringTasks)
	r.POST("/recurring-tasks/register", createRecurringTask)

	slog.Info("Task service starting", "port", 8085)
	r.Run(":8085")
}

func loadRecurringTasks(c *cron.Cron) {
	var tasks []models.RecurringTask
	db.Where("is_active = ?", true).Find(&tasks)
	for _, t := range tasks {
		task := t // Capture loop variable
		_, err := c.AddFunc(t.Schedule, func() {
			triggerRecurringTask(&task)
		})
		if err != nil {
			slog.Error("Failed to schedule recurring task", "task", t.Name, "err", err)
		}
	}
}

func triggerRecurringTask(rt *models.RecurringTask) {
	ctx, span := otel.Tracer("task").Start(context.Background(), "TriggerRecurringTask")
	defer span.End()

	slog.Info("Triggering recurring task", "name", rt.Name, "endpoint", rt.Endpoint)

	req, _ := http.NewRequestWithContext(ctx, "POST", rt.Endpoint, bytes.NewBuffer([]byte(rt.Payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := otelClient.Do(req)
	if err != nil {
		slog.Error("Failed to trigger recurring task", "name", rt.Name, "err", err)
		return
	}
	defer resp.Body.Close()

	now := time.Now()
	db.Model(rt).Update("last_run_at", &now)
	slog.Info("Recurring task triggered", "name", rt.Name, "status", resp.StatusCode)
}
