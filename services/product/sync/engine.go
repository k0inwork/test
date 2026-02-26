package sync

import (
	"log"
	"pum-go/pkg/models"
	"pum-go/services/product/mock"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type SyncEngine struct {
	DB   *gorm.DB
	GLPI *mock.GLPIProvider
}

func NewSyncEngine(db *gorm.DB) *SyncEngine {
	return &SyncEngine{
		DB:   db,
		GLPI: &mock.GLPIProvider{},
	}
}

// ParseName parses the GLPI name format: REGION/#seq-POU name
func ParseName(fullName string) (region string, seq int, pouType string, name string) {
	// Example: MSK/1-ПОУ Rack 1
	parts := strings.Split(fullName, "/")
	if len(parts) < 2 {
		return "", 0, "", fullName
	}
	region = parts[0]

	rest := strings.SplitN(parts[1], "-", 2)
	if len(rest) < 2 {
		return region, 0, "", parts[1]
	}
	seq, _ = strconv.Atoi(rest[0])

	typeAndName := strings.SplitN(rest[1], " ", 2)
	pouType = typeAndName[0]
	if len(typeAndName) > 1 {
		name = typeAndName[1]
	}

	return
}

func (e *SyncEngine) Run() error {
	log.Println("Starting synchronization with mock GLPI...")
	assets := e.GLPI.GetAssets()

	for _, asset := range assets {
		region, seq, pouType, name := ParseName(asset.Name)

		var product models.Product
		// 1. Search by Index (Name, Region, SeqNum)
		result := e.DB.Where("name = ? AND region = ? AND sequential_number = ?", name, region, seq).First(&product)

		if result.Error == gorm.ErrRecordNotFound {
			// 2. Search by Foreign ID (GLPI UUID) - Rename Detection
			result = e.DB.Where("glpi_uuid = ?", asset.ID).First(&product)
			if result.Error == nil {
				log.Printf("Rename detected: %s -> %s\n", product.Name, name)
			} else {
				log.Printf("Creating new node: %s\n", name)
				product = models.Product{GlpiUUID: asset.ID}
			}
		}

		// Update fields
		product.Name = name
		product.Region = region
		product.SequentialNumber = seq
		product.PouType = pouType
		product.Long = asset.Long
		product.Lat = asset.Lat
		product.Address = asset.Address
		product.Geo = asset.Lat + ";" + asset.Long
		product.GlpiUUID = asset.ID

		if err := e.DB.Save(&product).Error; err != nil {
			log.Printf("Failed to save product %s: %v\n", name, err)
		}
	}

	log.Println("Synchronization complete.")
	return nil
}
