package actions

import (
	"context"
	"fmt"
)

// Action defines a system operation that can be triggered by any UI.
type Action interface {
	ID() string
	Description() string
	Execute(ctx context.Context, params map[string]interface{}) error
}

// Registry maintains a list of all available system actions.
type Registry struct {
	actions map[string]Action
}

func NewRegistry() *Registry {
	return &Registry{actions: make(map[string]Action)}
}

func (r *Registry) Register(a Action) {
	r.actions[a.ID()] = a
}

func (r *Registry) Get(id string) (Action, bool) {
	a, ok := r.actions[id]
	return a, ok
}

// Example Action: Reboot
type RebootAction struct{}

func (a RebootAction) ID() string          { return "node.reboot" }
func (a RebootAction) Description() string { return "Reboot a physical node" }
func (a RebootAction) Execute(ctx context.Context, params map[string]interface{}) error {
	nodeID := params["node_id"]
	fmt.Printf("Executing Reboot for node: %v\n", nodeID)
	// Logic to call Product/Hardware microservices
	return nil
}
