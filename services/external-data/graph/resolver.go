// Package graph provides the GraphQL resolver implementation for the
// external-data microservice, binding GraphQL operations to backend Go logic.
package graph

import "pum-go/pkg/external"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Provider external.Provider
}
