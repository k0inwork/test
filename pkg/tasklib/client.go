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
)

var (
	taskServiceURL string
	httpClient     *http.Client
)

// Init initializes the task library with the URL of the central task microservice.
func Init(url string) {
	taskServiceURL = url
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
}

// Spawn registers a new task, executes the given function in a goroutine, and
// automatically manages the state machine and alarm lifecycle based on the function's result.
func Spawn(ctx context.Context, username, operation, objectID, className string, fn func() error) (uint, error) {
	if taskServiceURL == "" {
		return 0, fmt.Errorf("tasklib not initialized")
	}

	taskID, err := createTask(username, operation, objectID, className)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}

	go runTask(taskID, objectID, className, fn)

	return taskID, nil
}

func runTask(taskID uint, objectID, className string, fn func() error) {
	defer func() {
		if r := recover(); r != nil {
			errMessage := fmt.Sprintf("panic: %v", r)
			slog.Error("Task panicked", "task_id", taskID, "error", r)
			_ = updateTaskStatus(taskID, models.TaskStateError, errMessage)
		}
	}()

	if err := updateTaskStatus(taskID, models.TaskStatePending, ""); err != nil {
		slog.Error("Failed to update task to pending", "task_id", taskID, "error", err)
	}

	err := fn()

	if err == nil {
		_ = updateTaskStatus(taskID, models.TaskStateFinished, "")
		if objectID != "" && className != "" {
			closeActiveAlarms(objectID, className, taskID)
		}
		return
	}

	var resErr *ResourceConnectionError
	if errors.As(err, &resErr) {
		_ = updateTaskStatus(taskID, models.TaskStateError, err.Error())
		raiseAlarm(taskID, objectID, className, err.Error())
		return
	}

	_ = updateTaskStatus(taskID, models.TaskStateFailed, err.Error())
}

func createTask(username, operation, objectID, className string) (uint, error) {
	req := map[string]string{
		"username":   username,
		"operation":  operation,
		"object_id":  objectID,
		"class_name": className,
	}
	body, _ := json.Marshal(req)

	resp, err := postWithRetry(fmt.Sprintf("%s/api/tasks", taskServiceURL), body)
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

func updateTaskStatus(taskID uint, status, resultMessage string) error {
	req := map[string]string{
		"status":         status,
		"result_message": resultMessage,
	}
	body, _ := json.Marshal(req)

	reqURL := fmt.Sprintf("%s/api/tasks/%d/status", taskServiceURL, taskID)

	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		reqObj, _ := http.NewRequest(http.MethodPatch, reqURL, bytes.NewBuffer(body))
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

func raiseAlarm(taskID uint, objectID, className, message string) {
	req := map[string]interface{}{
		"task_id":    taskID,
		"object_id":  objectID,
		"class_name": className,
		"message":    message,
	}
	body, _ := json.Marshal(req)

	resp, err := postWithRetry(fmt.Sprintf("%s/api/alarms", taskServiceURL), body)
	if err != nil {
		slog.Error("Failed to raise alarm", "task_id", taskID, "error", err)
		return
	}
	defer resp.Body.Close()
}

func closeActiveAlarms(objectID, className string, closedByTaskID uint) {
	req := map[string]interface{}{
		"object_id":         objectID,
		"class_name":        className,
		"closed_by_task_id": closedByTaskID,
	}
	body, _ := json.Marshal(req)

	resp, err := postWithRetry(fmt.Sprintf("%s/api/alarms/close-active", taskServiceURL), body)
	if err != nil {
		slog.Error("Failed to close active alarms", "object_id", objectID, "error", err)
		return
	}
	defer resp.Body.Close()
}

func postWithRetry(url string, body []byte) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		resp, err = httpClient.Post(url, "application/json", bytes.NewBuffer(body))
		if err == nil {
			return resp, nil
		}
		time.Sleep(time.Duration(i*2) * time.Second)
	}
	return nil, fmt.Errorf("failed after retries: %w", err)
}
