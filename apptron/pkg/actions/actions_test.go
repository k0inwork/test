package actions

import (
	"context"
	"testing"
)

func TestRebootAction(t *testing.T) {
	action := RebootAction{}
	params := map[string]interface{}{"node_id": "test-node"}

	err := action.Execute(context.Background(), params)
	if err != nil {
		t.Errorf("RebootAction failed: %v", err)
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	a := RebootAction{}
	r.Register(a)

	got, ok := r.Get("node.reboot")
	if !ok {
		t.Fatal("Action not found in registry")
	}
	if got.ID() != "node.reboot" {
		t.Errorf("Expected node.reboot, got %s", got.ID())
	}
}
