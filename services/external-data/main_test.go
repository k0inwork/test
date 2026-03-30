// Package main contains integration and setup tests for the external-data
// microservice to verify routing and fundamental GraphQL operations.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pum-go/pkg/logging"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExternalDataAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logging.Init("test-external-data")
	r := setupRouter()

	// Test GET /problems
	req, _ := http.NewRequest("GET", "/problems", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var problems []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &problems)
	assert.Len(t, problems, 2)
	assert.Equal(t, "MSK-SW-01", problems[0]["node"])
	assert.Equal(t, "SPB-SW-02", problems[1]["node"])

	// Test GET / (GraphQL playground)
	req, _ = http.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test POST /query (GraphQL)
	query := `{"query":"{ pdus { id } }"}`
	req, _ = http.NewRequest("POST", "/query", bytes.NewBuffer([]byte(query)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var gqlResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &gqlResp)

	// Verify GraphQL data structure returned
	data, ok := gqlResp["data"].(map[string]interface{})
	assert.True(t, ok)
	assets, ok := data["pdus"].([]interface{})
	assert.True(t, ok)
	assert.Greater(t, len(assets), 0)
}
