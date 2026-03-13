package main

import (
	"bytes"
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

	// Mock external modules proxy server for tests
	mockExternalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer mockExternalServer.Close()

	// Override the external URL in the package variable
	externalModulesURL = mockExternalServer.URL

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

	// Test POST /subnets
	subnetPayload := SubnetReq{
		ID:      "123",
		Pools:   "10.0.0.10-10.0.0.50,10.0.0.100-10.0.0.200",
		Subnet:  "10.0.0.0/24",
		Relay:   "192.168.1.1,192.168.1.2",
		Options: "domain-name-servers=8.8.8.8,routers=10.0.0.1",
	}
	payloadBytes, _ := json.Marshal(subnetPayload)
	req, _ = http.NewRequest("POST", "/subnets", bytes.NewBuffer(payloadBytes))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test DELETE /subnets
	subnetDeletePayload := SubnetReq{ID: "123"}
	payloadBytes, _ = json.Marshal(subnetDeletePayload)
	req, _ = http.NewRequest("DELETE", "/subnets", bytes.NewBuffer(payloadBytes))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)


	// Test GET /dhcp
	req, _ = http.NewRequest("GET", "/dhcp", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var dhcpResp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &dhcpResp)
	assert.Len(t, dhcpResp, 1)

	// Test POST /dhcp
	dhcpPayload := DHCPReq{
		Hostname: "test-pc",
		Address:  "10.10.1.55",
		Mac:      "00:11:22:33:44:55",
		SubnetID: "1",
	}
	payloadBytes, _ = json.Marshal(dhcpPayload)
	req, _ = http.NewRequest("POST", "/dhcp", bytes.NewBuffer(payloadBytes))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test DELETE /dhcp
	dhcpDeletePayload := DHCPReq{
		Hostname: "test-pc",
		Address:  "10.10.1.55",
	}
	payloadBytes, _ = json.Marshal(dhcpDeletePayload)
	req, _ = http.NewRequest("DELETE", "/dhcp", bytes.NewBuffer(payloadBytes))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)


	// Test GET /dns
	req, _ = http.NewRequest("GET", "/dns", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var dnsResp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &dnsResp)
	assert.Len(t, dnsResp, 2)

	// Test POST /dns
	dnsPayload := DNSReq{
		Hostname: "server2.local",
		Address:  "10.10.1.101",
	}
	payloadBytes, _ = json.Marshal(dnsPayload)
	req, _ = http.NewRequest("POST", "/dns", bytes.NewBuffer(payloadBytes))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test DELETE /dns
	dnsDeletePayload := DNSReq{
		Hostname: "server2.local",
		Address:  "10.10.1.101",
	}
	payloadBytes, _ = json.Marshal(dnsDeletePayload)
	req, _ = http.NewRequest("DELETE", "/dns", bytes.NewBuffer(payloadBytes))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

}
