# The Admin Command Center: Apptron-based NMS Client

## Vision
The goal is to provide network administrators with a unified, local-first "Command Center" that combines the best of all worlds:
- **VSCode-based Editor**: For writing scripts, configuration templates, and automation workflows.
- **Interactive Dashboards**: High-level visual overviews of network status, traffic, and health.
- **Powerful CLI/TUI**: Fast, keyboard-driven management tools for power users who prefer the terminal over a mouse.

By leveraging **Apptron**, we create an environment that runs entirely in the browser but feels like a local Linux machine.

## Key Principles
1. **CLI First**: Every GUI action must be reproducible via a CLI command. This enables scripting and automation.
2. **TUI for Interactivity**: For complex but frequent tasks, a Text User Interface (TUI) provides a low-latency, focus-oriented alternative to traditional web forms.
3. **WASM-Powered**: Both the TUI and the "Headless Core" (backing the web dashboards) run as WebAssembly, ensuring high performance and security.
4. **Local-First, Cloud-Connected**: The client runs locally in the browser, persisting data via IndexedDB, while connecting to the main PUM-Go microservices for global state and device control.

## User Workflow
1. **Monitor**: Open a Dashboard to see real-time alerts.
2. **Diagnose**: Switch to the TUI (integrated in the terminal) to drill down into a specific switch or gateway.
3. **Remediate**: Write a quick script in the Editor using the CLI tools to batch-update configurations across multiple devices.
4. **Automate**: Save that script to the local persistent filesystem for future use.
