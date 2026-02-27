package sync

import (
	"log/slog"
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
	slog.Info("Starting synchronization", "provider", "GLPI (Mock)")
	assets := e.GLPI.GetAssets()
	slog.Debug("Fetched assets from provider", "count", len(assets))

	for _, asset := range assets {
		region, seq, pouType, name := ParseName(asset.Name)

		slog.Debug("Processing asset",
			"glpi_id", asset.ID,
			"full_name", asset.Name,
			"parsed_name", name,
			"region", region,
			"seq", seq,
		)

		var product models.Product
		// 1. Search by Index (Name, Region, SeqNum)
		result := e.DB.Where("name = ? AND region = ? AND sequential_number = ?", name, region, seq).First(&product)

		if result.Error == gorm.ErrRecordNotFound {
			slog.Debug("Node not found by name index, checking for rename", "glpi_id", asset.ID)

			// 2. Search by Foreign ID (GLPI UUID) - Rename Detection
			result = e.DB.Where("glpi_uuid = ?", asset.ID).First(&product)
			if result.Error == nil {
				slog.Info("Rename detected",
					"glpi_id", asset.ID,
					"old_name", product.Name,
					"new_name", name,
					"region", region,
				)
			} else {
				slog.Info("New node discovered", "name", name, "glpi_id", asset.ID)
				product = models.Product{GlpiUUID: asset.ID}
			}
		} else {
			slog.Debug("Matching node found in database", "id", product.ID, "name", product.Name)
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
			slog.Error("Failed to save product", "name", name, "error", err)
		} else {
			slog.Debug("Product updated successfully", "id", product.ID, "name", product.Name)
		}
	}

	slog.Info("Synchronization complete")
	return nil
}
