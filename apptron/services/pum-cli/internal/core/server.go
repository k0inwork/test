// +build js,wasm

package core

import (
	"syscall/js"
)

// StartHeadlessServer exposes PUM logic to the browser via JS globals.
func StartHeadlessServer() {
	js.Global().Set("pum", map[string]interface{}{
		"getStatus": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			return "PUM Core: Running"
		}),
		"fetchNodes": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// In a real implementation, this would call the PUM-Go microservices
			return []interface{}{"Node-01", "Node-02"}
		}),
	})

	// Keep the WASM instance alive
	select {}
}
