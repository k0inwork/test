package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pum-go/pkg/logging"
	"pum-go/pkg/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Subnet{}, &models.IPAddress{})
	return db
}

func TestNetworkAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init("test-network")
	testDB := setupTestDB()
	r := setupRouter(testDB)

	// Test GET /subnets (empty)
	req, _ := http.NewRequest("GET", "/subnets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var subnets []models.Subnet
	json.Unmarshal(w.Body.Bytes(), &subnets)
	assert.Len(t, subnets, 0)

	// Test GET /subnets (with data)
	testDB.Create(&models.Subnet{Prefix: "10.0.0.0/24", Region: "test-region"})
	req, _ = http.NewRequest("GET", "/subnets", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &subnets)
	assert.Len(t, subnets, 1)
	assert.Equal(t, "10.0.0.0/24", subnets[0].Prefix)

	// Test GET /dhcp
	req, _ = http.NewRequest("GET", "/dhcp", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var dhcpResp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &dhcpResp)
	assert.Len(t, dhcpResp, 1)

	// Test GET /dns
	req, _ = http.NewRequest("GET", "/dns", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var dnsResp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &dnsResp)
	assert.Len(t, dnsResp, 2)
}
