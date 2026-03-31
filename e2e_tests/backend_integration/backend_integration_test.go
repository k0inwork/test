package backend_integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const ExternalModulesURL = "http://localhost:8086"

func waitForService(url string) bool {
	for i := 0; i < 5; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func TestExternalModulesProxy(t *testing.T) {
	if !waitForService("http://localhost:8088/discovery") {
		handleMissingEnv(t, "Skipping integration test: Registry not available (run start.sh first)")
		return
	}

	if !waitForService(ExternalModulesURL) {
		handleMissingEnv(t, "Skipping integration test: External Modules proxy not available")
		return
	}

	reqBody := map[string]interface{}{
		"target_ip": "10.0.0.5",
		"command":   "dhcp host add",
		"param":     "{\"hostname\": \"test-host\"}",
	}

	jsonValue, _ := json.Marshal(reqBody)
	resp, err := http.Post(fmt.Sprintf("%s/call", ExternalModulesURL), "application/json", bytes.NewBuffer(jsonValue))

	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	assert.Equal(t, "success", res["status"])
	assert.Contains(t, res["response"], "dhcp host add")
}

func TestPDUCommandProxy(t *testing.T) {
	if !waitForService("http://localhost:8088/discovery") {
		handleMissingEnv(t, "Skipping integration test: Registry not available (run start.sh first)")
		return
	}

	if !waitForService(ExternalModulesURL) {
		handleMissingEnv(t, "Skipping integration test: External Modules proxy not available")
		return
	}

	reqBody := map[string]interface{}{
		"target_ip": "10.0.0.10",
		"command":   "restart",
		"param":     "1",
	}

	jsonValue, _ := json.Marshal(reqBody)
	resp, err := http.Post(fmt.Sprintf("%s/pdu/relay", ExternalModulesURL), "application/json", bytes.NewBuffer(jsonValue))

	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	assert.Equal(t, "success", res["status"])
	assert.Contains(t, res["message"], "PDU Command restart executed")
}
