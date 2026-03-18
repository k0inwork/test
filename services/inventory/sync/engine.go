package sync

import (
	"context"
	"fmt"
	"log/slog"
	"pum-go/pkg/external"
	"pum-go/pkg/models"

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

func (e *SyncEngine) Run() error {
	slog.Info("Starting inventory synchronization")

	// 1. Sync Switches
	equipments, err := e.Provider.GetNetworkEquipment(context.Background())
	if err != nil {
		slog.Error("Failed to fetch network equipment", "error", err)
	} else {
		for _, eq := range equipments {
			var sw models.Switch
			result := e.DB.Where("glpi_id = ?", eq.ID).First(&sw)
			if result.Error == gorm.ErrRecordNotFound {
				slog.Info("New switch discovered", "name", eq.Name, "glpi_id", eq.ID)
				sw = models.Switch{ID: eq.ID, GlpiID: eq.ID}
			}

			sw.Name = eq.Name
			sw.IP = eq.IP
			sw.Model = eq.Model
			sw.Status = eq.Status
			sw.Serial = eq.Serial
			sw.Manufacturer = eq.Manufacturer
			sw.Firmware = eq.Firmware
			sw.PortsCount = 48

			if err := e.DB.Save(&sw).Error; err != nil {
				slog.Error("Failed to save switch", "name", sw.Name, "error", err)
				continue
			}

			// Update or Create ports
			for i := 1; i <= 3; i++ {
				portID := fmt.Sprintf("p-%s-%d", sw.ID, i)
				var port models.SwitchPort
				if err := e.DB.Where("id = ?", portID).First(&port).Error; err == gorm.ErrRecordNotFound {
					port = models.SwitchPort{
						ID:       portID,
						SwitchID: sw.ID,
						Vlan:     10,
					}
				}
				port.Port = fmt.Sprintf("%s:Port %d", sw.Name, i)
				e.DB.Save(&port)
			}
		}
	}

	// 2. Sync IPMI (Computers)
	computers, err := e.Provider.GetComputers(context.Background())
	if err != nil {
		slog.Error("Failed to fetch computers (IPMI)", "error", err)
	} else {
		for _, c := range computers {
			var ipmi models.Ipmi
			result := e.DB.Where("id = ?", c.ID).First(&ipmi)
			if result.Error == gorm.ErrRecordNotFound {
				slog.Info("New IPMI discovered", "name", c.Name, "id", c.ID)
				ipmi = models.Ipmi{ID: c.ID}
			}

			ipmi.Name = c.Name
			ipmi.IP = c.IP
			ipmi.Status = c.Status
			ipmi.Dns = c.DNS
			ipmi.Available = (c.Status == "Online" || c.Status == "1")

			if err := e.DB.Save(&ipmi).Error; err != nil {
				slog.Error("Failed to save IPMI", "name", ipmi.Name, "error", err)
			}
		}
	}

	// 3. Sync PDUs
	pdus, err := e.Provider.GetPDUs(context.Background())
	if err != nil {
		slog.Error("Failed to fetch PDUs", "error", err)
	} else {
		for _, p := range pdus {
			var pdu models.PDU
			result := e.DB.Where("glpi_id = ?", p.ID).First(&pdu)
			if result.Error == gorm.ErrRecordNotFound {
				slog.Info("New PDU discovered", "name", p.Name, "glpi_id", p.ID)
				pdu = models.PDU{ID: p.ID, GlpiID: p.ID}
			}

			pdu.Name = p.Name
			pdu.IP = "" // PDU IP might come from another source if not in product.json
			pdu.Model = p.Model
			pdu.Manufacturer = p.Manufacturer
			pdu.Serial = p.Serial

			if err := e.DB.Save(&pdu).Error; err != nil {
				slog.Error("Failed to save PDU", "name", pdu.Name, "error", err)
			}
		}
	}

	slog.Info("Inventory synchronization complete")
	return nil
}
