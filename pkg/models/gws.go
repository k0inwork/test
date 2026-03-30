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

type Session struct {
	ID           string `gorm:"primaryKey"`
	Name         string
	Gw1ID        string
	Gw2ID        string
	UserID       string
	Subnet       string
	Successful   int
	Unsuccessful int
	TxBytes      int
	TxBytesOld   int
}

func (s *Session) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return
}
