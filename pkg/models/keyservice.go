package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KeyService struct {
	ID                    string `gorm:"primaryKey"`
	Order                 string
	Gw1ID                 string
	Port1                 string
	Gw2ID                 string
	Port2                 string
	Client                string
	KeyLength             int
	KeyChange             int
	Status                string `gorm:"default:'INITIATED'"`
	ConnectionInformation string `gorm:"default:'1:1'"`
}

func (k *KeyService) BeforeCreate(tx *gorm.DB) (err error) {
	if k.ID == "" {
		k.ID = uuid.New().String()
	}
	return
}
