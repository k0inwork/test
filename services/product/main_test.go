package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	testDB := setupProductTestDB()
	db = testDB

	r := gin.Default()
	engine := sync.NewSyncEngine(db)

	r.GET("/nodes", func(c *gin.Context) {
		var products []models.Product
		db.Find(&products)
		c.JSON(http.StatusOK, products)
	})
	r.POST("/sync", func(c *gin.Context) {
		engine.Run()
		c.JSON(http.StatusOK, gin.H{"message": "Sync completed"})
	})

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
