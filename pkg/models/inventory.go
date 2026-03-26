// Package models defines the inventory-related GORM models, representing
// physical and logical assets within the system's database structure.
package models

import (
	"time"

	"gorm.io/gorm"
)

type Switch struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	GlpiID      string         `gorm:"size:20" json:"glpi_id"`
	NodeID      uint           `json:"node_id"` // Reference to Product/Node
	LogicalType string         `gorm:"size:2;default:'cl'" json:"logical_type"`
	PortsCount  int            `json:"ports_count"`
	IP          string         `gorm:"size:45;default:'0.0.0.0'" json:"ip"`
	Model       string         `gorm:"size:200" json:"model"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type SwitchPort struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	SwitchID    string         `gorm:"index" json:"switch_id"`
	Port        string         `gorm:"size:35;not null" json:"port"`
	Description string         `gorm:"size:1000" json:"description"`
	Vlan        int            `gorm:"default:0" json:"vlan"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
