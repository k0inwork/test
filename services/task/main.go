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

func main() {
	tp, _ := tracing.InitTracer("task")
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", "err", err)
		}
	}()
	logging.Init("task")
	initDB()

	// Start background jobs
	go startArchiverJob()
	go startSchedulerJob()

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "task",
		Endpoint: "http://localhost:8085",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "tasks", Endpoints: []string{"/tasks"}},
			{Name: "alarms", Endpoints: []string{"/alarms"}},
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

	api := r.Group("/api")
	{
		api.GET("/tasks", listTasks)
		api.POST("/tasks", createTask)
		api.PATCH("/tasks/:id/status", updateTaskStatus)

		api.GET("/alarms", listAlarms)
		api.POST("/alarms", createAlarm)
		api.PATCH("/alarms/:id/status", updateAlarmStatus)
		api.POST("/alarms/close-active", closeActiveAlarms)

		api.GET("/recurring", listRecurringTasks)
		api.POST("/recurring", createRecurringTask)
	}

	// Legacy endpoint for compatibility
	r.GET("/tasks", listTasks)

	slog.Info("Task service starting", "port", 8085)
	r.Run(":8085")
}

func listTasks(c *gin.Context) {
	var tasks []models.TaskRecord
	query := db.WithContext(c.Request.Context()).Order("started_at desc")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	query.Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

func createTask(c *gin.Context) {
	var req models.TaskRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Status = models.TaskStateInitiated
	req.StartedAt = time.Now()
	if err := db.WithContext(c.Request.Context()).Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func updateTaskStatus(c *gin.Context) {
	var req struct {
		Status        string `json:"status"`
		ResultMessage string `json:"result_message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var task models.TaskRecord
	if err := db.WithContext(c.Request.Context()).First(&task, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	task.Status = req.Status
	if req.ResultMessage != "" {
		task.ResultMessage = req.ResultMessage
	}
	if task.Status == models.TaskStateFinished || task.Status == models.TaskStateFailed || task.Status == models.TaskStateError {
		now := time.Now()
		task.FinishedAt = &now
	}
	db.WithContext(c.Request.Context()).Save(&task)
	c.JSON(http.StatusOK, task)
}

func listAlarms(c *gin.Context) {
	var alarms []models.Alarm
	db.WithContext(c.Request.Context()).Find(&alarms)
	c.JSON(http.StatusOK, alarms)
}

func createAlarm(c *gin.Context) {
	var req models.Alarm
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Status = models.AlarmStatusRaised
	db.WithContext(c.Request.Context()).Create(&req)
	c.JSON(http.StatusCreated, req)
}

func updateAlarmStatus(c *gin.Context) {
	var req struct { Status string `json:"status"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var alarm models.Alarm
	if err := db.WithContext(c.Request.Context()).First(&alarm, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alarm not found"})
		return
	}
	alarm.Status = req.Status
	db.WithContext(c.Request.Context()).Save(&alarm)
	c.JSON(http.StatusOK, alarm)
}

func closeActiveAlarms(c *gin.Context) {
	var req struct { ObjectID string `json:"object_id"`; ClassName string `json:"class_name"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.WithContext(c.Request.Context()).Model(&models.Alarm{}).
		Where("resource_id = ? AND service_name = ? AND status = ?", req.ObjectID, req.ClassName, models.AlarmStatusRaised).
		Update("status", models.AlarmStatusClosed)
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func listRecurringTasks(c *gin.Context) {
	var tasks []models.RecurringTask
	db.WithContext(c.Request.Context()).Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

func createRecurringTask(c *gin.Context) {
	var req models.RecurringTask
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.IsActive = true
	var existing models.RecurringTask
	if err := db.WithContext(c.Request.Context()).Where("name = ?", req.Name).First(&existing).Error; err == nil {
		req.ID = existing.ID
		db.WithContext(c.Request.Context()).Save(&req)
		c.JSON(http.StatusOK, req)
	} else {
		db.WithContext(c.Request.Context()).Create(&req)
		c.JSON(http.StatusCreated, req)
	}
}

func startArchiverJob() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		db.Model(&models.TaskRecord{}).
			Where("status IN ? AND updated_at < ?", []string{models.TaskStateFinished, models.TaskStateFailed, models.TaskStateError}, time.Now().Add(-24*time.Hour)).
			Update("status", models.TaskStateArchived)
	}
}

func startSchedulerJob() {
	c := cron.New()
	c.Start()
	activeJobs := make(map[uint]cron.EntryID)
	for {
		var tasks []models.RecurringTask
		db.Where("is_active = ?", true).Find(&tasks)
		seen := make(map[uint]bool)
		for _, t := range tasks {
			seen[t.ID] = true
			if _, exists := activeJobs[t.ID]; !exists {
				task := t
				id, _ := c.AddFunc(t.Schedule, func() {
					slog.Info("Triggering recurring task", "name", task.Name)
					ctxT, spanT := otel.Tracer("scheduler").Start(context.Background(), "TriggerRecurringTask:"+task.Name)
					defer spanT.End()
					req, _ := http.NewRequestWithContext(ctxT, "POST", task.TargetURL, bytes.NewBuffer([]byte(task.Payload)))
					req.Header.Set("Content-Type", "application/json")
					resp, err := otelClient.Do(req)
					if err == nil { resp.Body.Close() }
					now := time.Now()
					db.Model(&models.RecurringTask{}).Where("id = ?", task.ID).Update("last_run_at", &now)
				})
				activeJobs[t.ID] = id
			}
		}
		for id, entryID := range activeJobs {
			if !seen[id] {
				c.Remove(entryID)
				delete(activeJobs, id)
			}
		}
		time.Sleep(1 * time.Minute)
	}
}
