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

func TestIPValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Valid IPv4", "192.168.1.1", true},
		{"Valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"Valid IPv6 compressed", "::1", true},
		{"Invalid IP - hostname", "localhost", false},
		{"Invalid IP - domain", "example.com", false},
		{"Invalid IP - typo", "192.168.1.999", false},
		{"Invalid IP - injection", "192.168.1.1; ls", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isValidIP(tt.ip))
		})
	}
}

func TestCommandValidation(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected bool
	}{
		{"Valid simple command", "power_on", true},
		{"Valid simple command 2", "status", true},
		{"Valid command with hyphen", "hard-reset", true},
		{"Valid camel case", "statusRelayAll", true},
		{"Invalid command - spaces", "power on", false},
		{"Invalid command - injection semicolon", "status; rm -rf /", false},
		{"Invalid command - injection pipe", "status | grep IP", false},
		{"Invalid command - injection backticks", "status `ls`", false},
		{"Invalid command - path traversal", "../../../etc/passwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isValidCommand(tt.cmd))
		})
	}
}

func performRequest(r http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestEndpointsValidation(t *testing.T) {
	// Initialize logger to prevent nil pointer in middleware
	logging.Init("external-modules-test")
	r := setupRouter()

	endpoints := []string{"/call", "/pdu/relay", "/ipmi/power"}

	for _, endpoint := range endpoints {
		t.Run("Valid payload on "+endpoint, func(t *testing.T) {
			payload := ModuleRequest{
				TargetIP: "10.0.0.1",
				Command:  "status",
				Param:    "1",
			}
			body, _ := json.Marshal(payload)
			w := performRequest(r, "POST", endpoint, body)

			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("Invalid IP on "+endpoint, func(t *testing.T) {
			payload := ModuleRequest{
				TargetIP: "10.0.0.1; ls",
				Command:  "status",
				Param:    "1",
			}
			body, _ := json.Marshal(payload)
			w := performRequest(r, "POST", endpoint, body)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			var response map[string]string
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, "invalid target_ip format", response["error"])
		})

		t.Run("Invalid Command on "+endpoint, func(t *testing.T) {
			payload := ModuleRequest{
				TargetIP: "10.0.0.1",
				Command:  "status; cat /etc/passwd",
				Param:    "1",
			}
			body, _ := json.Marshal(payload)
			w := performRequest(r, "POST", endpoint, body)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			var response map[string]string
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, "invalid command format", response["error"])
		})
	}
}
