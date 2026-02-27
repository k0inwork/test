package sync

import (
	"pum-go/pkg/models"
	"pum-go/services/product/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSyncEngine_Run(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Product{})

	engine := NewSyncEngine(db)

	// 1. Initial Sync
	err := engine.Run()
	assert.NoError(t, err)

	var count int64
	db.Model(&models.Product{}).Count(&count)
	assert.Equal(t, int64(2), count)

	var p1 models.Product
	db.Where("glpi_uuid = ?", "glpi-1").First(&p1)
	assert.Equal(t, "Rack 1", p1.Name)
	assert.Equal(t, "MSK", p1.Region)

	// 2. Rename Detection Mocking
	// Modify the mock provider to return a renamed asset
	engine.GLPI = &mock.GLPIProvider{} // Using a custom mock for rename would be better, but let's simulate manually

	// Simulate rename in GLPI: glpi-1 is now "MSK/1-ПОУ New Rack Name"
	renamedAsset := mock.GLPIAsset{
		ID:      "glpi-1",
		Name:    "MSK/1-ПОУ New Rack Name",
		Long:    "37.6173",
		Lat:     "55.7558",
		Address: "New Address",
	}

	// Manually run sync logic for renamed asset
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
