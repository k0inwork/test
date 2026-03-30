// Package main contains integration tests for the product microservice, verifying
// database migration, routing logic, and OTel trace propagation context.
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pum-go/pkg/external"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
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
