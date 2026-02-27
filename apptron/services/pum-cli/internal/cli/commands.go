package cli

import (
	"context"
	"fmt"
	"net/http"
	"pum-go/apptron/pkg/actions"

	"github.com/spf13/cobra"
)

func AddActionCommands(root *cobra.Command, registry *actions.Registry) {
	// Reboot Action
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

	// Ping Bridge Action
	root.AddCommand(&cobra.Command{
		Use:   "ping-bridge",
		Short: "Test connectivity to the Bridge Agent",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Pinging Bridge Agent via Virtual Network...")
			// In live mode, this would be a real GET to the bridge's virtual IP
			resp, err := http.Get("http://10.0.0.1/health")
			if err != nil {
				fmt.Printf("Ping failed: %v (Expected in mock mode)\n", err)
				return
			}
			defer resp.Body.Close()
			fmt.Printf("Bridge Response: %d OK\n", resp.StatusCode)
		},
	})
}
