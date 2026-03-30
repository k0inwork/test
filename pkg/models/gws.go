// Package models defines data structures and GORM models for Gateway (gws)
// entities, facilitating database operations for gateway management.
package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Gw struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex:idx_name_region"`
	Region      string `gorm:"uniqueIndex:idx_name_region"`
	Description string
	Address     string
	State       string
	Log         string
}

func (g *Gw) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	return
}
