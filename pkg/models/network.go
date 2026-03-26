// Package models provides GORM definitions for network-related entities such as
// IP addresses, subnets, and VLANs used within the infrastructure management context.
package models

import (
	"time"

	"gorm.io/gorm"
)

type Subnet struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Prefix    string         `gorm:"size:45;not null" json:"prefix"` // e.g. 192.168.1.0/24
	Region    string         `gorm:"size:20" json:"region"`
	Type      string         `gorm:"size:20" json:"type"` // e.g. "management", "data"
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type IPAddress struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Address   string         `gorm:"size:45;not null" json:"address"`
	SubnetID  uint           `json:"subnet_id"`
	Status    string         `gorm:"size:20;default:'allocated'" json:"status"`
	Reserved  bool           `json:"reserved"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
