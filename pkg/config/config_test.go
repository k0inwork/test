package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Success(t *testing.T) {
	validYAML := []byte(`
system:
  name: "test-system"
  version: "1.0.0"
  env: "development"
core_services:
  - "auth"
  - "db"
optional_services:
  - "cache"
external_modules:
  module1:
    mode: "mock"
    endpoint: "http://localhost:8080"
    real_endpoint: "http://example.com"
discovery:
  registry_url: "http://registry:8500"
  heartbeat_interval: "10s"
  timeout: "5s"
`)

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	err := os.WriteFile(configPath, validYAML, 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "test-system", cfg.System.Name)
	assert.Equal(t, "1.0.0", cfg.System.Version)
	assert.Equal(t, "development", cfg.System.Env)
	assert.Equal(t, []string{"auth", "db"}, cfg.CoreServices)
	assert.Equal(t, []string{"cache"}, cfg.OptionalServices)
	assert.Len(t, cfg.ExternalModules, 1)
	assert.Equal(t, "mock", cfg.ExternalModules["module1"].Mode)
	assert.Equal(t, "http://localhost:8080", cfg.ExternalModules["module1"].Endpoint)
	assert.Equal(t, "http://example.com", cfg.ExternalModules["module1"].RealEndpoint)
	assert.Equal(t, "http://registry:8500", cfg.Discovery.RegistryURL)
	assert.Equal(t, "10s", cfg.Discovery.HeartbeatInterval)
	assert.Equal(t, "5s", cfg.Discovery.Timeout)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.yaml")

	cfg, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.True(t, os.IsNotExist(err))
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	invalidYAML := []byte(`
system:
  name: "test-system"
  version: "1.0.0"
  env: "development"
core_services:
  - "auth"
 - "db"  # invalid indentation
`)

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")
	err := os.WriteFile(configPath, invalidYAML, 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
