// Package models outlines the GORM models for products (nodes/servers) and
// associated hardware configurations managed by the platform.
package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Name             string         `gorm:"size:100;not null" json:"name"`
	Description      string         `gorm:"size:1000" json:"description"`
	Address          string         `gorm:"size:1000" json:"address"`
	State            string         `gorm:"size:1000;default:'Узел функционирует исправно'" json:"state"`
	Long             string         `gorm:"size:12;default:'30.0'" json:"long"`
	Lat              string         `gorm:"size:12;default:'60.0'" json:"lat"`
	Geo              string         `gorm:"size:100" json:"geo"`
	Operate          bool           `gorm:"default:true" json:"operate"`
	GlpiUUID         string         `gorm:"size:36" json:"glpi_uuid"`
	SequentialNumber int            `gorm:"default:1" json:"sequential_number"`
	PouType          string         `gorm:"size:10;default:'ПОУ'" json:"pou_type"`
	Region           string         `gorm:"size:20;default:''" json:"region"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
