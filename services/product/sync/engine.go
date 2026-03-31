// Package sync defines a synchronization engine for the product service to
// reconcile local product records with an external source of truth.
package sync

import (
	"context"
	"fmt"
	"log/slog"
	"pum-go/pkg/external"
	"pum-go/pkg/models"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type SyncEngine struct {
	DB       *gorm.DB
	Provider external.Provider
}

func NewSyncEngine(db *gorm.DB, provider external.Provider) *SyncEngine {
	return &SyncEngine{
		DB:       db,
		Provider: provider,
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

func (e *SyncEngine) Run(ctx context.Context) error {
	slog.Info("Starting synchronization", "provider", "External Data Service")
	assets, err := e.Provider.GetPDUs(ctx)
	if err != nil {
		slog.Error("Failed to fetch assets from provider", "error", err)
		return err
	}
	slog.Debug("Fetched assets from provider", "count", len(assets))

	// Fetch all existing products to avoid N+1 queries
	var existingProducts []models.Product
	if err := e.DB.WithContext(ctx).Find(&existingProducts).Error; err != nil {
		slog.Error("Failed to fetch existing products", "error", err)
		return err
	}

	// Create lookup maps
	byIndex := make(map[string]*models.Product)
	byUUID := make(map[string]*models.Product)
	for i := range existingProducts {
		p := &existingProducts[i]
		indexKey := fmt.Sprintf("%s|%d|%s", p.Region, p.SequentialNumber, p.Name)
		byIndex[indexKey] = p
		if p.GlpiUUID != "" {
			byUUID[p.GlpiUUID] = p
		}
	}

	var productsToSave []models.Product

	for _, asset := range assets {
		region, seq, pouType, name := ParseName(asset.Name)

		slog.Debug("Processing asset",
			"glpi_id", asset.ID,
			"full_name", asset.Name,
			"parsed_name", name,
			"region", region,
			"seq", seq,
		)

		var product *models.Product
		indexKey := fmt.Sprintf("%s|%d|%s", region, seq, name)

		// 1. Search by Index (Name, Region, SeqNum)
		if p, ok := byIndex[indexKey]; ok {
			product = p
			slog.Debug("Matching node found in database", "id", product.ID, "name", product.Name)
		} else {
			slog.Debug("Node not found by name index, checking for rename", "glpi_id", asset.ID)
			// 2. Search by Foreign ID (GLPI UUID) - Rename Detection
			if p, ok := byUUID[asset.ID]; ok {
				product = p
				slog.Info("Rename detected",
					"glpi_id", asset.ID,
					"old_name", product.Name,
					"new_name", name,
					"region", region,
				)
			} else {
				slog.Info("New node discovered", "name", name, "glpi_id", asset.ID)
				product = &models.Product{GlpiUUID: asset.ID}
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

		productsToSave = append(productsToSave, *product)
	}

	if len(productsToSave) > 0 {
		if err := e.DB.WithContext(ctx).Save(&productsToSave).Error; err != nil {
			slog.Error("Failed to save products in batch", "error", err)
			return err
		}
		slog.Debug("Products updated successfully in batch", "count", len(productsToSave))
	}

	slog.Info("Synchronization complete")
	return nil
}
