package tasklib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"pum-go/pkg/logging"
	"time"

	"github.com/gin-gonic/gin"
)

// RecurringTaskHandler is the function signature for recurring task callbacks
type RecurringTaskHandler func(ctx context.Context, payload []byte) error

// RegisterEndpoint registers a local webhook endpoint for recurring tasks and
// announces it to the central task microservice.
func RegisterEndpoint(registryURL string, r *gin.Engine, path, schedule, targetURL, username, operation, objectID, className string, handler RecurringTaskHandler) {
	// 1. Define the local webhook
	r.POST(path, func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
			return
		}

		slog.Info("Executing recurring task", "operation", operation, "object_id", objectID)
		if err := handler(c.Request.Context(), body); err != nil {
			slog.Error("Recurring task execution failed", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// 2. Register with Task service asynchronously
	go func() {
		// Wait for registry to be available
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		_ = logging.WaitForService(ctx, registryURL+"/discovery")

		// Register the recurring task
		reqBody := map[string]interface{}{
			"name":       fmt.Sprintf("%s:%s", className, operation),
			"schedule":   schedule,
			"target_url": targetURL,
			"payload":    "{}",
			"is_active":  true,
		}
		body, _ := json.Marshal(reqBody)

		for {
			if taskServiceURL == "" {
				time.Sleep(1 * time.Second)
				continue
			}
			resp, err := httpClient.Post(taskServiceURL+"/api/recurring", "application/json", bytes.NewBuffer(body))
			if err == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
				resp.Body.Close()
				slog.Info("Recurring task registered successfully", "operation", operation)
				break
			}
			if err != nil {
				slog.Warn("Failed to register recurring task, retrying...", "error", err)
			} else {
				slog.Warn("Failed to register recurring task, unexpected status", "status", resp.StatusCode)
				resp.Body.Close()
			}
			time.Sleep(10 * time.Second)
		}
	}()
}
