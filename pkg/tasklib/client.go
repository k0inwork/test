package tasklib

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"pum-go/pkg/models"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	taskServiceURL string
	httpClient     *http.Client
)

// Init initializes the task library with the URL of the central task microservice.
func Init(url string) {
	taskServiceURL = url
	httpClient = &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   30 * time.Second,
	}
}

// Spawn registers a new task, executes the given function in a goroutine, and
// automatically manages the state machine and alarm lifecycle based on the function's result.
func Spawn(ctx context.Context, username, operation, objectID, className string, fn func(context.Context) error) (uint, error) {
	if taskServiceURL == "" {
		return 0, fmt.Errorf("tasklib not initialized")
	}

	taskID, err := createTask(ctx, username, operation, objectID, className)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}

	go runTask(ctx, taskID, objectID, className, fn)

	return taskID, nil
}

func runTask(ctx context.Context, taskID uint, objectID, className string, fn func(context.Context) error) {
	defer func() {
		if r := recover(); r != nil {
			errMessage := fmt.Sprintf("panic: %v", r)
			slog.Error("Task panicked", "task_id", taskID, "error", r)
			_ = updateTaskStatus(ctx, taskID, models.TaskStateError, errMessage)
		}
	}()

	if err := updateTaskStatus(ctx, taskID, models.TaskStatePending, ""); err != nil {
		slog.Error("Failed to update task to pending", "task_id", taskID, "error", err)
	}

	err := fn(ctx)

	if err == nil {
		_ = updateTaskStatus(ctx, taskID, models.TaskStateFinished, "")
		if objectID != "" && className != "" {
			closeActiveAlarms(ctx, objectID, className, taskID)
		}
		return
	}

	var resErr *ResourceConnectionError
	if errors.As(err, &resErr) {
		_ = updateTaskStatus(ctx, taskID, models.TaskStateError, err.Error())
		raiseAlarm(ctx, taskID, objectID, className, err.Error())
		return
	}

	_ = updateTaskStatus(ctx, taskID, models.TaskStateFailed, err.Error())
}

func createTask(ctx context.Context, username, operation, objectID, className string) (uint, error) {
	req := map[string]string{
		"username":   username,
		"operation":  operation,
		"object_id":  objectID,
		"class_name": className,
	}
	body, _ := json.Marshal(req)

	resp, err := postWithRetry(ctx, fmt.Sprintf("%s/api/tasks", taskServiceURL), body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res models.TaskRecord
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, err
	}

	return res.ID, nil
}

func updateTaskStatus(ctx context.Context, taskID uint, status, resultMessage string) error {
	req := map[string]string{
		"status":         status,
		"result_message": resultMessage,
	}
	body, _ := json.Marshal(req)

	reqURL := fmt.Sprintf("%s/api/tasks/%d/status", taskServiceURL, taskID)

	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		reqObj, _ := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, bytes.NewBuffer(body))
		reqObj.Header.Set("Content-Type", "application/json")

		resp, err = httpClient.Do(reqObj)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i*2) * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func raiseAlarm(ctx context.Context, taskID uint, objectID, className, message string) {
	req := map[string]interface{}{
		"task_id":    taskID,
		"object_id":  objectID,
		"class_name": className,
		"message":    message,
	}
	body, _ := json.Marshal(req)

	resp, err := postWithRetry(ctx, fmt.Sprintf("%s/api/alarms", taskServiceURL), body)
	if err != nil {
		slog.Error("Failed to raise alarm", "task_id", taskID, "error", err)
		return
	}
	defer resp.Body.Close()
}

func closeActiveAlarms(ctx context.Context, objectID, className string, closedByTaskID uint) {
	req := map[string]interface{}{
		"object_id":         objectID,
		"class_name":        className,
		"closed_by_task_id": closedByTaskID,
	}
	body, _ := json.Marshal(req)

	resp, err := postWithRetry(ctx, fmt.Sprintf("%s/api/alarms/close-active", taskServiceURL), body)
	if err != nil {
		slog.Error("Failed to close active alarms", "object_id", objectID, "error", err)
		return
	}
	defer resp.Body.Close()
}

func postWithRetry(ctx context.Context, url string, body []byte) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err = httpClient.Do(req)
		if err == nil {
			return resp, nil
		}
		time.Sleep(time.Duration(i*2) * time.Second)
	}
	return nil, fmt.Errorf("failed after retries: %w", err)
}
