package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pum-go/pkg/external"
	"pum-go/pkg/models"
	"pum-go/pkg/logging"
	"pum-go/services/product/sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProductTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Product{})
	return db
}

func TestProductAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init("test-product")
	testDB := setupProductTestDB()

	provider := &external.MockProvider{}
	engine := sync.NewSyncEngine(testDB, provider)
	r := setupRouter(testDB, engine)

	// Test GET /nodes (Empty)
	req, _ := http.NewRequest("GET", "/nodes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 0)

	// Test POST /sync
	req, _ = http.NewRequest("POST", "/sync", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test GET /nodes (Populated)
	req, _ = http.NewRequest("GET", "/nodes", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Greater(t, len(products), 0)
}

func TestGatewayAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init("test-product-gw")
	testDB := setupProductTestDB()

	engine := sync.NewSyncEngine(testDB, nil)
	r := setupRouter(testDB, engine)

	// Test POST /gateways
	gw := models.Product{
		Name:    "Test GW",
		Region:  "MSK",
		Address: "1.1.1.1",
		Log:     "Initial Log",
	}
	body, _ := json.Marshal(gw)
	req, _ := http.NewRequest("POST", "/gateways", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createdGW models.Product
	json.Unmarshal(w.Body.Bytes(), &createdGW)
	assert.Equal(t, "GW", createdGW.PouType)
	assert.Equal(t, "Test GW", createdGW.Name)

	// Test GET /gateways
	req, _ = http.NewRequest("GET", "/gateways", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var gateways []models.Product
	json.Unmarshal(w.Body.Bytes(), &gateways)
	assert.Len(t, gateways, 1)
	assert.Equal(t, "Test GW", gateways[0].Name)
}
