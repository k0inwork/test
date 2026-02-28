package sync

import (
	"pum-go/pkg/external"
	"pum-go/pkg/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInventorySyncEngine_Run(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Switch{}, &models.SwitchPort{})

	provider := &external.MockProvider{}
	engine := NewSyncEngine(db, provider)

	// 1. Initial Sync
	err := engine.Run()
	assert.NoError(t, err)

	var swCount int64
	db.Model(&models.Switch{}).Count(&swCount)
	assert.Equal(t, int64(2), swCount)

	var portCount int64
	db.Model(&models.SwitchPort{}).Count(&portCount)
	assert.Equal(t, int64(6), portCount) // 2 switches * 3 ports each

	var sw1 models.Switch
	db.Where("glpi_id = ?", "sw-1").First(&sw1)
	assert.Equal(t, "MSK-SW-01", sw1.Name)
}
