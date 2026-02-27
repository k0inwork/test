package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pum-go/pkg/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.User{})
	return db
}

func TestUserAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	testDB := setupTestDB()
	db = testDB

	r := gin.Default()
	r.GET("/users", func(c *gin.Context) {
		var users []models.User
		db.Find(&users)
		c.JSON(http.StatusOK, users)
	})
	r.POST("/users", func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Create(&user)
		c.JSON(http.StatusOK, user)
	})

	// Test POST /users
	user := models.User{Username: "testuser", Role: "admin"}
	jsonValue, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var createdUser models.User
	json.Unmarshal(w.Body.Bytes(), &createdUser)
	assert.Equal(t, "testuser", createdUser.Username)

	// Test GET /users
	req, _ = http.NewRequest("GET", "/users", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var users []models.User
	json.Unmarshal(w.Body.Bytes(), &users)
	assert.Len(t, users, 1)
	assert.Equal(t, "testuser", users[0].Username)
}
