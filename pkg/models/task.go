package models

import (
	"time"

	"gorm.io/gorm"
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
