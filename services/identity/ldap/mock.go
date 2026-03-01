package ldap

import (
	"log/slog"
)

type MockLDAPProvider struct{}

func NewMockLDAPProvider() *MockLDAPProvider {
	return &MockLDAPProvider{}
}

func (m *MockLDAPProvider) Authenticate(username, password string) (bool, string, error) {
	slog.Info("Authenticating via mock LDAP", "username", username)
	if username == "admin" && password == "admin" {
		return true, "admin", nil
	}
	return true, "user", nil
}
