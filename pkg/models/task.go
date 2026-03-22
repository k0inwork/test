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
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Status    string         `json:"status"` // INITIATED, PENDING, FINISHED, FAILED, ERROR, ARCHIVED
	Progress  int            `json:"progress"`
	Payload   string         `json:"payload"`
	Result    string         `json:"result"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Alarm struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ResourceID  string         `gorm:"index" json:"resource_id"`
	ServiceName string         `json:"service_name"`
	Severity    string         `json:"severity"` // CRITICAL, WARNING, INFO
	Message     string         `json:"message"`
	Active      bool           `gorm:"default:true" json:"active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type RecurringTask struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Schedule  string         `json:"schedule"`   // e.g. "@every 5m" or cron expression
	Endpoint  string         `json:"endpoint"`   // The webhook URL to call
	Payload   string         `json:"payload"`    // JSON payload to send
	IsActive  bool           `gorm:"default:true" json:"is_active"` // Whether the task is currently scheduled
	LastRunAt *time.Time     `json:"last_run_at"` // Track the last execution
	NextRunAt *time.Time     `json:"next_run_at"` // Optional: used by some schedulers
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
