package sync

import (
	"context"
	"fmt"
	"pum-go/pkg/external"
	"pum-go/pkg/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type BenchmarkMockProvider struct {
	Assets []*external.Gpdu
}

func (p *BenchmarkMockProvider) GetNetworkEquipment(ctx context.Context) ([]*external.GNetworkEquipment, error) {
	return nil, nil
}
func (p *BenchmarkMockProvider) GetPDUs(ctx context.Context) ([]*external.Gpdu, error) {
	return p.Assets, nil
}
func (p *BenchmarkMockProvider) GetComputers(ctx context.Context) ([]*external.GComputer, error) {
	return nil, nil
}
func (p *BenchmarkMockProvider) GetHosts(ctx context.Context) ([]*external.ZHost, error) {
	return nil, nil
}

func BenchmarkRun(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	db.AutoMigrate(&models.Product{})

	// Pre-fill 50 products
	for i := 0; i < 50; i++ {
		db.Create(&models.Product{
			Name:             fmt.Sprintf("Rack %d", i),
			Region:           "MSK",
			SequentialNumber: i,
			GlpiUUID:         fmt.Sprintf("uuid-%d", i),
		})
	}

	// Create 100 assets (50 updates, 50 new)
	assets := make([]*external.Gpdu, 100)
	for i := 0; i < 100; i++ {
		assets[i] = &external.Gpdu{
			ID:      fmt.Sprintf("uuid-%d", i),
			Name:    fmt.Sprintf("MSK/%d-ПОУ Rack %d", i, i),
			Long:    "37.6173",
			Lat:     "55.7558",
			Address: "Moscow",
			Model:   "APC",
		}
	}

	provider := &BenchmarkMockProvider{Assets: assets}
	engine := NewSyncEngine(db, provider)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.Run(context.Background())
	}
}
