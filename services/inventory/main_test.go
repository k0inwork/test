// Package main contains tests verifying the startup and routing configurations
// of the inventory microservice.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pum-go/pkg/external"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"pum-go/pkg/tasklib"
	"pum-go/services/inventory/sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pum-go/pkg/config"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Switch{}, &models.SwitchPort{}, &models.Ipmi{}, &models.PDU{})
	return db
}

func TestInventoryAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init("test-inventory")
	tasklib.Init("http://localhost:8085")
	testDB := setupTestDB()

	provider := &external.MockProvider{}
	engine := sync.NewSyncEngine(testDB, provider)
	r := setupRouter(testDB, engine)

	// Test GET /switches
	testDB.Create(&models.Switch{ID: "sw1", Name: "test-switch", IP: "10.0.0.1"})
	req, _ := http.NewRequest("GET", "/switches", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var switches []models.Switch
	json.Unmarshal(w.Body.Bytes(), &switches)
	assert.Len(t, switches, 1)
	assert.Equal(t, "test-switch", switches[0].Name)

	// Test GET /pdus
	req, _ = http.NewRequest("GET", "/pdus", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test GET /ipmi
	req, _ = http.NewRequest("GET", "/ipmi", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test POST /configurable
	cfg := config.Config{
		ExternalModules: map[string]struct {
			Mode         string `yaml:"mode"`
			Endpoint     string `yaml:"endpoint"`
			RealEndpoint string `yaml:"real_endpoint"`
		}{
			"test-module": {Mode: "mock"},
		},
	}
	jsonValue, _ := json.Marshal(cfg)
	req, _ = http.NewRequest("POST", "/configurable", bytes.NewBuffer(jsonValue))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
