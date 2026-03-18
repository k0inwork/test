package sync

import (
	"context"
	"pum-go/pkg/external"
	"pum-go/pkg/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSyncEngine(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Switch{}, &models.SwitchPort{})

	provider := &external.MockProvider{}
	engine := NewSyncEngine(db, provider)

	err := engine.Run(context.Background())
	assert.NoError(t, err)
}
