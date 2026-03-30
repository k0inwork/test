// Package external contains unit tests for the generic external provider
// interface and any common provider-related logic.
package external

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMockProvider_GetPDUs(t *testing.T) {
	p := &MockProvider{}
	assets, err := p.GetPDUs(context.Background())
	assert.NoError(t, err)
	assert.Len(t, assets, 2)
	assert.Equal(t, "MSK/1-ПОУ Rack 1", assets[0].Name)
}

func TestMockProvider_GetHosts(t *testing.T) {
	p := &MockProvider{}
	hosts, err := p.GetHosts(context.Background())
	assert.NoError(t, err)
	assert.Len(t, hosts, 2)
	assert.Equal(t, "MSK-SW-01", hosts[0].Name)
	assert.NotEmpty(t, hosts[0].Problems)
}
