package models

import (
	"time"

	"github.com/google/uuid"
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
	Log              string         `gorm:"size:1000" json:"log"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type Gw struct {
	ID          string `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"uniqueIndex:idx_name_region" json:"name"`
	Region      string `gorm:"uniqueIndex:idx_name_region" json:"region"`
	Description string `json:"description"`
	Address     string `json:"address"`
	State       string `json:"state"`
	Log         string `json:"log"`
}

func (g *Gw) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	return
}

type Session struct {
	ID           string `gorm:"primaryKey" json:"id"`
	Name         string `json:"name"`
	Gw1ID        string `json:"gw1_id"`
	Gw2ID        string `json:"gw2_id"`
	UserID       string `json:"user_id"`
	Subnet       string `json:"subnet"`
	Successful   int    `json:"successful"`
	Unsuccessful int    `json:"unsuccessful"`
	TxBytes      int    `json:"tx_bytes"`
	TxBytesOld   int    `json:"tx_bytes_old"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return
}
