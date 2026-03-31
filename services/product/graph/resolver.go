// Package graph acts as the root GraphQL resolver for the product service,
// enabling data fetching for products, nodes, and gateways.
package graph

import (
	"gorm.io/gorm"
	"pum-go/pkg/external"
	"pum-go/services/product/sync"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB       *gorm.DB
	Sync     *sync.SyncEngine
	Provider external.Provider
}
