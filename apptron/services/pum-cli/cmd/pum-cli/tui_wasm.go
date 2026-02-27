// +build js

package main

import (
	"fmt"
)

func startTUI() {
	fmt.Println("TUI mode is currently disabled in WASM (Phase 1).")
	fmt.Println("Please use CLI commands (e.g., 'pum status', 'pum reboot').")
}
