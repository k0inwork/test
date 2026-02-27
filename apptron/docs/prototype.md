# Prototype Structure: PUM WASM Client

This prototype demonstrates how a single Go codebase can support CLI, TUI, and Headless Server modes when compiled to WASM for Apptron.

## Project Structure
```
services/pum-cli/
├── cmd/
│   └── pum-cli/
│       └── main.go       # Entry point with mode detection
├── internal/
│   ├── tui/              # Bubbletea programs
│   ├── cli/              # Cobra commands
│   └── core/             # WASM Headless logic (syscall/js)
└── pkg/
    └── client/           # API client for main microservices
```

## Mode Detection (main.go)
The tool detects its environment using environment variables and JS global checks.

```go
func main() {
    if isWasmServer() {
        startHeadlessServer()
    } else {
        executeCobra()
    }
}

func isWasmServer() bool {
    // Check if running in a WebWorker context or via a specific flag
    return os.Getenv("PUM_MODE") == "headless"
}
```

## Headless Server (Go + syscall/js)
Exposes Go functions to the browser's JavaScript environment.

```go
func startHeadlessServer() {
    js.Global().Set("pumFetchNodes", js.FuncOf(func(this js.Value, args []js.Value) any {
        // Fetch nodes from microservices and return to JS
        return nil
    }))
    select {} // Keep WASM alive
}
```

## CLI/TUI (pum-cli)
Uses Cobra for CLI and Bubbletea for TUI.

```go
// In Cobra Run:
if interactive {
    p := tea.NewProgram(tui.InitialModel())
    p.Run()
} else {
    cli.RunCommand()
}
```
