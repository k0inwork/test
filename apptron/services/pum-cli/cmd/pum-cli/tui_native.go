// +build !js

package main

import (
	"fmt"
	"os"
	"pum-go/apptron/services/pum-cli/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func startTUI() {
	p := tea.NewProgram(tui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
