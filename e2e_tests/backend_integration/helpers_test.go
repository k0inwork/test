package backend_integration

import (
	"os"
	"testing"
)

func handleMissingEnv(t *testing.T, msg string) {
	if os.Getenv("CI") == "true" {
		t.Fatal(msg)
	} else {
		t.Skip(msg)
	}
}
