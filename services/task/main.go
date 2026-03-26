// Package main executes the task microservice, serving as the central hub for
// asynchronous job states, resource alarms, and recurring scheduling tasks.
package main

import (
	"bytes"
	"context"
	"fmt"
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
	"gorm.io/gorm/logger"
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

	logging.RegisterWithDiscovery("http://localhost:8088", logging.ServiceRegistration{
		Name:     "task",
		Endpoint: "http://localhost:8085",
		Capabilities: []logging.CapabilityRegistration{
			{Name: "tasks", Endpoints: []string{"/tasks"}},
			{Name: "async-executor", Endpoints: []string{"/async-executor"}},
		},
		IsCore:  false,
		OrderID: 5,
		Menu: []logging.MenuItem{
			{Label: "Tasks", Path: "/tasks"},
		},
	})

	// Start background jobs
	go startArchiverJob()
	go startSchedulerJob()

	r := gin.Default()
	r.Use(otelgin.Middleware("task"))
	r.Use(logging.GinMiddleware())

	api := r.Group("/api")
	{
		// Task Endpoints
		api.GET("/tasks", listTasks)
		api.POST("/tasks", createTask)
		api.PATCH("/tasks/:id/status", updateTaskStatus)

		// Alarm Endpoints
		api.GET("/alarms", listAlarms)
		api.POST("/alarms", createAlarm)
		api.PATCH("/alarms/:id/status", updateAlarmStatus)
		api.POST("/alarms/close-active", closeActiveAlarms)

		// Recurring Task Endpoints
		api.GET("/recurring", listRecurringTasks)
		api.POST("/recurring", createRecurringTask)
	}

	// Legacy endpoint for backward compatibility (used by the menu manifest)
	r.GET("/tasks", listTasks)

	slog.Info("Task service starting", "port", 8085)
	r.Run(":8085")
}

// ---- Tasks ----

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

	// State Machine Validation
	validTransitions := map[string][]string{
		models.TaskStateInitiated: {models.TaskStatePending},
		models.TaskStatePending:   {models.TaskStateFinished, models.TaskStateFailed, models.TaskStateError},
		models.TaskStateFinished:  {models.TaskStateArchived},
		models.TaskStateFailed:    {models.TaskStateArchived},
		models.TaskStateError:     {models.TaskStateArchived},
		models.TaskStateArchived:  {}, // Terminal
	}

	allowedTransitions := validTransitions[task.Status]
	isValidTransition := false
	for _, allowed := range allowedTransitions {
		if req.Status == allowed {
			isValidTransition = true
			break
		}
	}

	if !isValidTransition {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid transition from %s to %s", task.Status, req.Status)})
		return
	}

	// Update Task
	task.Status = req.Status
	if req.ResultMessage != "" {
		task.ResultMessage = req.ResultMessage
	}

	// Handle terminal states completion time
	if task.Status == models.TaskStateFinished || task.Status == models.TaskStateFailed || task.Status == models.TaskStateError {
		now := time.Now()
		task.FinishedAt = &now
	}

	if err := db.WithContext(c.Request.Context()).Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// ---- Alarms ----

func listAlarms(c *gin.Context) {
	var alarms []models.Alarm
	query := db.WithContext(c.Request.Context()).Order("created_at desc")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if objID := c.Query("object_id"); objID != "" {
		query = query.Where("object_id = ?", objID)
	}
	if className := c.Query("class_name"); className != "" {
		query = query.Where("class_name = ?", className)
	}
	query.Find(&alarms)
	c.JSON(http.StatusOK, alarms)
}

func createAlarm(c *gin.Context) {
	var req models.Alarm
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Status = models.AlarmStatusRaised
	if err := db.WithContext(c.Request.Context()).Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

func updateAlarmStatus(c *gin.Context) {
	var req struct {
		Status         string `json:"status"`
		ClosedByTaskID *uint  `json:"closed_by_task_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var alarm models.Alarm
	if err := db.WithContext(c.Request.Context()).First(&alarm, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alarm not found"})
		return
	}

	if req.Status == models.AlarmStatusClosed {
		alarm.Status = models.AlarmStatusClosed
		alarm.ClosedByTaskID = req.ClosedByTaskID
	} else if req.Status == models.AlarmStatusRaised {
		alarm.Status = models.AlarmStatusRaised
		alarm.ClosedByTaskID = nil
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status for alarm"})
		return
	}

	if err := db.WithContext(c.Request.Context()).Save(&alarm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alarm)
}

func closeActiveAlarms(c *gin.Context) {
	var req struct {
		ObjectID       string `json:"object_id"`
		ClassName      string `json:"class_name"`
		ClosedByTaskID *uint  `json:"closed_by_task_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ObjectID == "" || req.ClassName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "object_id and class_name are required"})
		return
	}

	result := db.WithContext(c.Request.Context()).Model(&models.Alarm{}).
		Where("object_id = ? AND class_name = ? AND status = ?", req.ObjectID, req.ClassName, models.AlarmStatusRaised).
		Updates(map[string]interface{}{
			"status":            models.AlarmStatusClosed,
			"closed_by_task_id": req.ClosedByTaskID,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"closed_count": result.RowsAffected})
}

// ---- Recurring Tasks ----

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
	// Upsert based on Name to avoid duplicates when services restart
	var existing models.RecurringTask
	if err := db.WithContext(c.Request.Context()).Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).Where("name = ?", req.Name).First(&existing).Error; err == nil {
		existing.Schedule = req.Schedule
		existing.TargetURL = req.TargetURL
		existing.Payload = req.Payload
		existing.IsActive = true
		db.WithContext(c.Request.Context()).Save(&existing)
		c.JSON(http.StatusOK, existing)
		return
	}

	if err := db.WithContext(c.Request.Context()).Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// ---- Background Jobs ----

func startArchiverJob() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		<-ticker.C
		ctx, span := otel.Tracer("scheduler").Start(context.Background(), "ArchiveOldTasks")
		slog.Info("Running task archiver job")
		// Archive tasks older than 24 hours in terminal states
		threshold := time.Now().Add(-24 * time.Hour)
		result := db.WithContext(ctx).Model(&models.TaskRecord{}).
			Where("status IN ? AND updated_at < ?", []string{models.TaskStateFinished, models.TaskStateFailed, models.TaskStateError}, threshold).
			Update("status", models.TaskStateArchived)

		if result.Error != nil {
			slog.Error("Error archiving tasks", "error", result.Error)
		} else if result.RowsAffected > 0 {
			slog.Info("Archived old tasks", "count", result.RowsAffected)
		}
		span.End()
	}
}

// startSchedulerJob uses robfig/cron to manage the schedule of recurring tasks.
// Since tasks can be added dynamically, it runs a sync loop to ensure the cron runner
// is up to date with the database state.
func startSchedulerJob() {
	// Enable second-level precision to support "@every 10s" for testing
	// and finer-grained control.
	c := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))
	c.Start()

	// Map to keep track of active jobs so we don't duplicate them
	activeJobs := make(map[uint]cron.EntryID)

	ticker := time.NewTicker(1 * time.Minute)
	for {
		ctx, span := otel.Tracer("scheduler").Start(context.Background(), "SyncRecurringTasks")
		var tasks []models.RecurringTask
		if err := db.WithContext(ctx).Where("is_active = ?", true).Find(&tasks).Error; err != nil {
			slog.Error("Error fetching recurring tasks", "error", err)
			span.End()
			time.Sleep(1 * time.Minute)
			continue
		}

		// Keep track of tasks we saw in this loop to remove deleted/deactivated ones
		seenTasks := make(map[uint]bool)

		for _, task := range tasks {
			seenTasks[task.ID] = true
			if _, exists := activeJobs[task.ID]; !exists {
				// We need to add this new/updated task to the cron scheduler
				t := task // capture loop variable
				entryID, err := c.AddFunc(t.Schedule, func() {
					ctxT, spanT := otel.Tracer("scheduler").Start(context.Background(), "TriggerRecurringTask:"+t.Name)
					defer spanT.End()

					slog.Info("Triggering recurring task", "name", t.Name, "url", t.TargetURL)
					req, _ := http.NewRequestWithContext(ctxT, "POST", t.TargetURL, bytes.NewBuffer([]byte(t.Payload)))
					req.Header.Set("Content-Type", "application/json")
					resp, err := otelClient.Do(req)
					if err != nil {
						slog.Error("Failed to trigger recurring task via webhook", "task_id", t.ID, "url", t.TargetURL, "error", err)
						return
					}
					defer resp.Body.Close()
					slog.Info("Recurring task webhook response", "task_id", t.ID, "status", resp.StatusCode)

					now := time.Now()
					db.WithContext(ctxT).Model(&models.RecurringTask{}).Where("id = ?", t.ID).Update("last_run_at", now)
				})

				if err != nil {
					slog.Error("Failed to add recurring task to cron scheduler", "task_id", t.ID, "schedule", t.Schedule, "error", err)
				} else {
					slog.Info("Added recurring task to cron scheduler", "task_id", t.ID, "schedule", t.Schedule)
					activeJobs[t.ID] = entryID
				}
			}
		}

		// Remove jobs from cron that are no longer active in the DB
		for taskID, entryID := range activeJobs {
			if !seenTasks[taskID] {
				c.Remove(entryID)
				delete(activeJobs, taskID)
				slog.Info("Removed inactive recurring task from cron scheduler", "task_id", taskID)
			}
		}
		span.End()

		<-ticker.C
	}
}
