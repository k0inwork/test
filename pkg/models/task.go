package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	TaskStateInitiated = "INITIATED"
	TaskStatePending   = "PENDING"
	TaskStateFinished  = "FINISHED"
	TaskStateFailed    = "FAILED"
	TaskStateError     = "ERROR"
	TaskStateArchived  = "ARCHIVED"
)

type TaskRecord struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Username      string         `json:"username"`
	Operation     string         `json:"operation"`
	Status        string         `json:"status"`
	ResultMessage string         `json:"result_message"`
	StartedAt     time.Time      `json:"started_at"`
	FinishedAt    *time.Time     `json:"finished_at"`
	ObjectID      string         `json:"object_id"`
	ClassName     string         `json:"class_name"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

const (
	AlarmStatusRaised = "RAISED"
	AlarmStatusClosed = "CLOSED"
)

type Alarm struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	TaskID          uint           `json:"task_id"` // ID of the task that created this alarm
	ClosedByTaskID  *uint          `json:"closed_by_task_id"`
	ObjectID        string         `json:"object_id"`
	ClassName       string         `json:"class_name"`
	Status          string         `json:"status"` // RAISED or CLOSED
	Message         string         `json:"message"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type RecurringTask struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `json:"name"`
	Schedule    string         `json:"schedule"`    // e.g. "@every 5m" or cron expression
	TargetURL   string         `json:"target_url"`  // The webhook URL to call
	Payload     string         `json:"payload"`     // JSON payload to send
	IsActive    bool           `json:"is_active"`   // Whether the task is currently scheduled
	LastRunAt   *time.Time     `json:"last_run_at"` // Track the last execution
	NextRunAt   *time.Time     `json:"next_run_at"` // Optional: used by some schedulers
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
