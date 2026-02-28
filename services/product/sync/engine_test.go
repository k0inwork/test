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
	db.AutoMigrate(&models.Product{})

	provider := &external.MockProvider{}
	engine := NewSyncEngine(db, provider)

	// 1. Initial Sync
	err := engine.Run()
	assert.NoError(t, err)

	var count int64
	db.Model(&models.Product{}).Count(&count)
	assert.Equal(t, int64(2), count)

	var p1 models.Product
	// In MockProvider, the PDU ID is "pdu-1" and name "MSK/1-ПОУ Rack 1"
	db.Where("glpi_uuid = ?", "pdu-1").First(&p1)
	assert.Equal(t, "Rack 1", p1.Name)
	assert.Equal(t, "MSK", p1.Region)

	// 2. Rename Detection Test
	// Simulate rename: glpi-1 is now "MSK/1-ПОУ New Rack Name"
	renamedAsset := external.Gpdu{
		ID:      "pdu-1",
		Name:    "MSK/1-ПОУ New Rack Name",
		Long:    "37.6173",
		Lat:     "55.7558",
		Address: "New Address",
	}

	region, seq, _, name := ParseName(renamedAsset.Name)
	var product models.Product
	// Search by Index (Old Name) - Should fail
	result := db.Where("name = ? AND region = ? AND sequential_number = ?", name, region, seq).First(&product)
	assert.Error(t, result.Error)

	// Search by Foreign ID (GLPI UUID) - Should find p1
	result = db.Where("glpi_uuid = ?", renamedAsset.ID).First(&product)
	assert.NoError(t, result.Error)
	assert.Equal(t, p1.ID, product.ID)

	// Update and save
	product.Name = name
	db.Save(&product)

	var p1Updated models.Product
	db.First(&p1Updated, p1.ID)
	assert.Equal(t, "New Rack Name", p1Updated.Name)
}
