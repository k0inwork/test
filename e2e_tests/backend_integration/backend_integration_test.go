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

// This integration test assumes that all services (including external-modules and registry)
// are either running via start.sh or docker-compose, or can be queried directly via localhost ports.
// Let's create a test that verifies making a proxy request to the external modules.

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
	// Let's skip if the service isn't running so it doesn't fail pure go test ./... locally
	// without the environment booted.
	if !waitForService("http://localhost:8088/services") {
		t.Skip("Skipping integration test: Registry not available (run start.sh first)")
	}

	if !waitForService(ExternalModulesURL) {
		t.Skip("Skipping integration test: External Modules proxy not available")
	}

	// Payload matching what external-modules expects
	reqBody := map[string]interface{}{
		"target_ip": "10.0.0.5",
		"command":   "dhcp host add",
		"param":     "{\"hostname\": \"test-host\"}",
	}

	jsonValue, _ := json.Marshal(reqBody)
	resp, err := http.Post(fmt.Sprintf("%s/call", ExternalModulesURL), "application/json", bytes.NewBuffer(jsonValue))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	assert.Equal(t, "success", res["status"])
	assert.Contains(t, res["response"], "dhcp host add")
}

func TestPDUCommandProxy(t *testing.T) {
	if !waitForService("http://localhost:8088/services") {
		t.Skip("Skipping integration test: Registry not available (run start.sh first)")
	}

	reqBody := map[string]interface{}{
		"target_ip": "10.0.0.10",
		"command":   "restart",
		"param":     "1",
	}

	jsonValue, _ := json.Marshal(reqBody)
	resp, err := http.Post(fmt.Sprintf("%s/pdu/relay", ExternalModulesURL), "application/json", bytes.NewBuffer(jsonValue))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	assert.Equal(t, "success", res["status"])
	assert.Contains(t, res["message"], "PDU Command restart executed")
}
