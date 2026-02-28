package main

import (
	"fmt"
	"os"
	"pum-go/apptron/pkg/actions"
	pumcli "pum-go/apptron/services/pum-cli/internal/cli"

	"github.com/spf13/cobra"
)

func main() {
	registry := actions.NewRegistry()
	registry.Register(actions.RebootAction{})

	var rootCmd = &cobra.Command{
		Use:   "pum",
		Short: "PUM Admin CLI/TUI",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				startTUI()
			}
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Get network status",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("PUM Network: OK")
			fmt.Println("Nodes: 42")
		},
	})

	pumcli.AddActionCommands(rootCmd, registry)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
