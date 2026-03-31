// Package graph holds the root resolver for the identity service's GraphQL API,
// managing access to underlying user and permission models.
package graph

import "gorm.io/gorm"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB *gorm.DB
}
