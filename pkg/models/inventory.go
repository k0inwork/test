// Package models defines the inventory-related GORM models, representing
// physical and logical assets within the system's database structure.
package models

import (
	"time"

	"gorm.io/gorm"
)

type Switch struct {
	ID           string         `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	GlpiID       string         `gorm:"size:20" json:"glpi_id"`
	NodeID       uint           `json:"node_id"` // Reference to Product/Node
	LogicalType  string         `gorm:"size:2;default:'cl'" json:"logical_type"`
	PortsCount   int            `json:"ports_count"`
	IP           string         `gorm:"size:45;default:'0.0.0.0'" json:"ip"`
	Model        string         `gorm:"size:200" json:"model"`
	Status       string         `gorm:"size:100" json:"status"`
	Serial       string         `gorm:"size:100" json:"serial"`
	Manufacturer string         `gorm:"size:100" json:"manufacturer"`
	Firmware     string         `gorm:"size:100" json:"firmware"`
	MacAddress   string         `gorm:"size:17" json:"mac_address"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
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

type Ipmi struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	IP          string         `gorm:"size:45" json:"ip"`
	Port        string         `gorm:"size:10;default:'623'" json:"port"`
	Status      string         `gorm:"size:100" json:"status"`
	Available   bool           `json:"available"`
	InterfaceID string         `json:"interface_id"`
	HostID      string         `json:"host_id"`
	UseIP       bool           `json:"use_ip"`
	Dns         string         `json:"dns"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type PDU struct {
	ID           string         `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	GlpiID       string         `gorm:"size:20" json:"glpi_id"`
	NodeID       uint           `json:"node_id"`
	IP           string         `gorm:"size:45" json:"ip"`
	Model        string         `gorm:"size:200" json:"model"`
	Manufacturer string         `gorm:"size:100" json:"manufacturer"`
	Serial       string         `gorm:"size:100" json:"serial"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
