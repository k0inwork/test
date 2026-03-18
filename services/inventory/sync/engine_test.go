package sync

import (
	"pum-go/pkg/external"
	"pum-go/pkg/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSyncEngine_Run(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Switch{}, &models.SwitchPort{}, &models.Ipmi{}, &models.PDU{})

	provider := &external.MockProvider{}
	engine := NewSyncEngine(db, provider)

	err := engine.Run()
	assert.NoError(t, err)

	var switches []models.Switch
	db.Find(&switches)
	assert.Len(t, switches, 2)
	assert.Equal(t, "MSK-SW-01", switches[0].Name)

	var ipmis []models.Ipmi
	db.Find(&ipmis)
	assert.Len(t, ipmis, 1)
	assert.Equal(t, "Server-01", ipmis[0].Name)

	var pdus []models.PDU
	db.Find(&pdus)
	assert.Len(t, pdus, 2)
	assert.Equal(t, "MSK/1-ПОУ Rack 1", pdus[0].Name)
}
