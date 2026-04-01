package tasklib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	expectedURL := "http://example.com/api/tasks"

	// Call Init
	Init(expectedURL)

	// Verify taskServiceURL is set correctly
	assert.Equal(t, expectedURL, taskServiceURL, "taskServiceURL should match the expected URL")

	// Verify httpClient is initialized
	assert.NotNil(t, httpClient, "httpClient should not be nil")

	// Verify httpClient properties
	if httpClient != nil {
		assert.Equal(t, 30*time.Second, httpClient.Timeout, "httpClient Timeout should be 30 seconds")
		assert.NotNil(t, httpClient.Transport, "httpClient Transport should not be nil")
	}
}
