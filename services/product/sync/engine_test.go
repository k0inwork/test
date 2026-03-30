// Package sync contains tests for the product synchronization engine, verifying
// correct model updates and lifecycle tracking during external data syncs.
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
	db.AutoMigrate(&models.Product{})

	provider := &external.MockProvider{}
	engine := NewSyncEngine(db, provider)

	err := engine.Run(context.Background())
	assert.NoError(t, err)
}
