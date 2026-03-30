// Package sync provides background synchronization engines for the inventory
// service to map and pull external data into the local database models.
package sync

import (
	"context"
	"log/slog"
	"pum-go/pkg/external"
	"pum-go/pkg/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (e *SyncEngine) Run(ctx context.Context) error {
	slog.Info("Starting inventory synchronization")
	equipments, err := e.Provider.GetNetworkEquipment(ctx)
	if err != nil {
		slog.Error("Failed to fetch network equipment", "error", err)
		return err
	}

	for _, eq := range equipments {
		var sw models.Switch
		result := e.DB.WithContext(ctx).Where("glpi_id = ?", eq.ID).First(&sw)
		if result.Error == gorm.ErrRecordNotFound {
			slog.Info("New switch discovered", "name", eq.Name, "glpi_id", eq.ID)
			// Use GLPI ID as internal ID if it's new
			sw = models.Switch{ID: eq.ID, GlpiID: eq.ID}
		}

		sw.Name = eq.Name
		sw.IP = eq.IP
		sw.Model = eq.Model
		sw.PortsCount = 48

		if err := e.DB.WithContext(ctx).Save(&sw).Error; err != nil {
			slog.Error("Failed to save switch", "name", sw.Name, "error", err)
			continue
		}

		// Update or Create ports
		var ports []models.SwitchPort
		for _, p := range eq.Ports {
			ports = append(ports, models.SwitchPort{
				ID:          p.ID,
				SwitchID:    sw.ID,
				Port:        p.Name,
				Description: p.Description,
				Vlan:        p.Vlan,
			})
		}

		if len(ports) > 0 {
			e.DB.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&ports)
		}
	}

	slog.Info("Inventory synchronization complete")
	return nil
}
