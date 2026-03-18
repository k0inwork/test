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
	db.AutoMigrate(&models.Product{}, &models.Gw{}, &models.Session{})
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

	r := setupRouter(testDB, nil)

	// Test POST /gateways
	gw := models.Gw{
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

	var createdGW models.Gw
	json.Unmarshal(w.Body.Bytes(), &createdGW)
	assert.NotEmpty(t, createdGW.ID)
	assert.Equal(t, "Test GW", createdGW.Name)

	// Test GET /gateways
	req, _ = http.NewRequest("GET", "/gateways", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var gateways []models.Gw
	json.Unmarshal(w.Body.Bytes(), &gateways)
	assert.Len(t, gateways, 1)
	assert.Equal(t, "Test GW", gateways[0].Name)
}

func TestSessionAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init("test-product-session")
	testDB := setupProductTestDB()

	r := setupRouter(testDB, nil)

	// Seed a session directly
	session := models.Session{
		Name: "Test Session",
		Subnet: "10.0.0.0/24",
	}
	testDB.Create(&session)

	// Test GET /sessions
	req, _ := http.NewRequest("GET", "/sessions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var sessions []models.Session
	json.Unmarshal(w.Body.Bytes(), &sessions)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "Test Session", sessions[0].Name)
}
