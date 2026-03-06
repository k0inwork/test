package models

import (
	"time"

	"gorm.io/gorm"
)

type Group struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"uniqueIndex;not null" json:"name"`
	Capabilities string         `json:"capabilities"` // Comma-separated list of capabilities
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Users        []User         `gorm:"many2many:user_groups;" json:"users,omitempty"`
}

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"not null" json:"-"`
	Role      string         `gorm:"default:user" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Groups    []Group        `gorm:"many2many:user_groups;" json:"groups,omitempty"`
}

// ActivityLog represents an audit log entry previously handled by Django's activity_log
type ActivityLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	RequestMethod string    `json:"request_method"`
	RequestURL    string    `json:"request_url"`
	QueryParams   string    `json:"query_params"`
	ResponseCode  int       `json:"response_code"`
	Datetime      time.Time `gorm:"autoCreateTime" json:"datetime"`
}
