package tasklib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"

	"github.com/gin-gonic/gin"
)

// RegisterEndpoint registers a reverse endpoint for the task microservice scheduler to call.
// It also spawns a goroutine to wait for the task service to be available and automatically
// registers its schedule.
func RegisterEndpoint(registryURL string, router *gin.Engine, path, schedule, targetURL, username, operation, objectID, className string, handler func(payload []byte) error) {
	// Setup the webhook endpoint to receive triggers
	router.POST(path, func(c *gin.Context) {
		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read payload"})
			return
		}

		// Spawn the task using the standard tasklib flow
		taskID, err := Spawn(context.Background(), username, operation, objectID, className, func() error {
			return handler(payload)
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to spawn task"})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"status":  "accepted",
			"task_id": taskID,
		})
	})

	// Background orchestration: Wait for task service and register schedule
	go func() {
		// Wait for the task microservice to be marked active in the central registry
		logging.WaitForService(registryURL, "task")

		// Register the recurring task
		reqBody := map[string]interface{}{
			"name":       fmt.Sprintf("%s:%s:%s", operation, objectID, className),
			"schedule":   schedule,
			"target_url": targetURL,
			"payload":    "{}", // Optional initial payload
		}

		bodyBytes, _ := json.Marshal(reqBody)
		if taskServiceURL == "" {
			slog.Error("Cannot register schedule: tasklib not initialized with Init()")
			return
		}

		resp, err := postWithRetry(taskServiceURL+"/api/recurring", bodyBytes)
		if err != nil {
			slog.Error("Failed to register recurring task with task service", "error", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			slog.Info("Successfully registered recurring task schedule", "url", targetURL)
		} else {
			slog.Error("Unexpected response registering recurring task", "status", resp.StatusCode)
		}
	}()
}
