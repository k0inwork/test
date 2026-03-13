package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pum-go/pkg/models"
	"pum-go/services/identity/ldap"
	"testing"

	"pum-go/pkg/logging"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.User{}, &models.Group{}, &models.ActivityLog{})
	return db
}

func TestUserAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init("test-identity") // Initialize logger to prevent nil pointer dereference
	testDB := setupTestDB()
	ldapMock = ldap.NewMockLDAPProvider()
	r := setupRouter(testDB)

	// Test GET /users (Empty)
	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test POST /login (Success)
	loginReq := map[string]string{"username": "admin", "password": "adminadmin"}
	jsonValue, _ := json.Marshal(loginReq)
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonValue))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	assert.Equal(t, "admin", loginResp["username"])
	assert.Equal(t, "logged_in", loginResp["status"])

	// Test GET /users (Should have admin now)
	req, _ = http.NewRequest("GET", "/users", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var usersResp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &usersResp)
	assert.Len(t, usersResp, 1)
	assert.Equal(t, "admin", usersResp[0]["username"])
}

func TestGroupAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init("test-identity")
	testDB := setupTestDB()
	r := setupRouter(testDB)

	testDB.Create(&models.User{Username: "testuser"})
	testDB.Create(&models.Group{Name: "testgroup"})

	// Test POST /users/:username/groups
	body := map[string]string{"group": "testgroup"}
	jsonValue, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/users/testuser/groups", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test GET /groups
	req, _ = http.NewRequest("GET", "/groups", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var groupsResp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &groupsResp)
	assert.Len(t, groupsResp, 1)
	assert.Equal(t, "testgroup", groupsResp[0]["name"])
	assert.Equal(t, "testuser", groupsResp[0]["users"])

	// Test DELETE /users/:username/groups/:group
	req, _ = http.NewRequest("DELETE", "/users/testuser/groups/testgroup", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify group is removed
	req, _ = http.NewRequest("GET", "/groups", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &groupsResp)
	assert.Equal(t, "", groupsResp[0]["users"])
}

func TestActivityListAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init("test-identity")
	testDB := setupTestDB()
	r := setupRouter(testDB)

	// Test POST /activitylist
	logEntry := models.ActivityLog{Username: "testuser", RequestMethod: "GET", RequestURL: "/test"}
	jsonValue, _ := json.Marshal(logEntry)
	req, _ := http.NewRequest("POST", "/activitylist", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test GET /activitylist
	req, _ = http.NewRequest("GET", "/activitylist", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var logsResp []models.ActivityLog
	json.Unmarshal(w.Body.Bytes(), &logsResp)
	assert.Len(t, logsResp, 1)
	assert.Equal(t, "testuser", logsResp[0].Username)
}
