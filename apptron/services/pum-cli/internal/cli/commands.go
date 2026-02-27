package cli

import (
	"context"
	"fmt"
	"pum-go/apptron/pkg/actions"

	"github.com/spf13/cobra"
)

func AddActionCommands(root *cobra.Command, registry *actions.Registry) {
	// Dynamically add a command for the "reboot" action
	if action, ok := registry.Get("node.reboot"); ok {
		cmd := &cobra.Command{
			Use:   "reboot [node_id]",
			Short: action.Description(),
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				params := map[string]interface{}{"node_id": args[0]}
				if err := action.Execute(context.Background(), params); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			},
		}
		root.AddCommand(cmd)
	}
}
