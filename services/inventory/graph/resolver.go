// Package graph implements the root GraphQL resolver for the inventory service,
// managing resolution context and dependencies for queries/mutations.
package graph

import (
	"gorm.io/gorm"
	"pum-go/services/inventory/sync"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB   *gorm.DB
	Sync *sync.SyncEngine
}
