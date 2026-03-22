package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pum-go/pkg/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() {
	var err error
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.TaskRecord{}, &models.Alarm{}, &models.RecurringTask{})
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	r.GET("/recurring", listRecurringTasks)
	r.POST("/recurring", createRecurringTask)
	return r
}

func TestRecurringTaskRegistration(t *testing.T) {
	setupTestDB()
	r := setupTestRouter()

	reqPayload := map[string]interface{}{
		"name":       "sync-switches",
		"schedule":   "@every 1m",
		"endpoint":   "http://localhost:8083/internal/tasks/sync",
		"payload":    "{}",
		"is_active":  true,
	}
	body, _ := json.Marshal(reqPayload)

	req, _ := http.NewRequest("POST", "/recurring", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var task models.RecurringTask
	db.First(&task)

	assert.Equal(t, "sync-switches", task.Name)
	assert.Equal(t, "@every 1m", task.Schedule)
	assert.Equal(t, "http://localhost:8083/internal/tasks/sync", task.Endpoint)
	assert.True(t, task.IsActive)

	// Test Upsert (duplicate name)
	reqPayload["endpoint"] = "http://localhost:9999/new-url"
	body, _ = json.Marshal(reqPayload)

	req2, _ := http.NewRequest("POST", "/recurring", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var taskUpdated models.RecurringTask
	db.First(&taskUpdated)

	assert.Equal(t, task.ID, taskUpdated.ID) // Should be same record
	assert.Equal(t, "http://localhost:9999/new-url", taskUpdated.Endpoint)
}

func TestCronSchedulerTrigger(t *testing.T) {
	setupTestDB()

	var webhookCalled bool
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalled = true
		w.WriteHeader(http.StatusAccepted)
	}))
	defer testServer.Close()

	// Register a task with a fast schedule for testing (e.g. 1 second)
	task := models.RecurringTask{
		Name:      "test-webhook",
		Schedule:  "@every 1s",
		Endpoint:  testServer.URL,
		Payload:   "{}",
		IsActive:  true,
	}
	db.Create(&task)

	// Trigger logic to verify
	triggerRecurringTask(&task)

	assert.True(t, webhookCalled, "Expected webhook to be called by the trigger logic")

	// Verify last_run_at was updated
	var updatedTask models.RecurringTask
	db.First(&updatedTask, task.ID)
	assert.NotNil(t, updatedTask.LastRunAt)
}
