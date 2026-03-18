package ldap

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
)

type MockLDAPProvider struct{}

func NewMockLDAPProvider() *MockLDAPProvider {
	return &MockLDAPProvider{}
}

// GroupInfo represents a mocked LDAP group and its assigned capabilities
type GroupInfo struct {
	Name         string
	Capabilities []string
}

func (m *MockLDAPProvider) Authenticate(ctx context.Context, username, password string) (bool, string, []GroupInfo, error) {
	_, span := otel.Tracer("MockLDAP").Start(ctx, "Authenticate")
	defer span.End()

	slog.Info("Authenticating via mock LDAP", "username", username)

	// Mock LDAP definitions
	adminGroup := GroupInfo{Name: "tsumadm", Capabilities: []string{"*"}}
	netGroup := GroupInfo{Name: "netadm", Capabilities: []string{"network", "ipam", "routing"}}
	devGroup := GroupInfo{Name: "devadm", Capabilities: []string{"inventory", "switches", "ports", "products"}}

	if username == "admin" && password == "admin" {
		return true, "admin", []GroupInfo{adminGroup}, nil
	}

	// For specific test users
	if username == "user_tsumadm" {
		return true, "admin", []GroupInfo{adminGroup}, nil
	}
	if username == "user_netadm" {
		return true, "user", []GroupInfo{netGroup}, nil
	}
	if username == "user_devadm" {
		return true, "user", []GroupInfo{devGroup}, nil
	}

	// Default to no capabilities
	return true, "user", []GroupInfo{}, nil
}
