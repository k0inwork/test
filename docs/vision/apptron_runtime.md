# The Apptron Runtime: A Platform for Network Management

## Overview
Apptron is a **runtime environment** that enables a "local-first" management experience within a web browser. While standard web templates are served via our Go frontend for basic access, Apptron provides the essential infrastructure to run complex, high-performance tools for power users.

## Core Runtime Components

### 1. WebViews for Interactive Dashboards
For data visualization and high-level monitoring, we use Apptron WebViews. These can be served directly from our Go-based microservices or the unified runner, providing:
-   Real-time network health visualization.
-   Interactive topology maps.
-   Integrated dashboards that feel like part of a native application.

### 2. WASM/WASI for CLI and TUI
The heart of the "Command Center" is the integration of Go-WASM. We compile our management tools (like `pum-cli`) into WebAssembly.
-   **CLI (Command Line Interface)**: Standard command-driven interaction (e.g., `pum node reboot`) running at near-native speed.
-   **TUI (Text User Interface)**: Rich, keyboard-interactive interfaces (built with frameworks like Bubbletea) that run inside a terminal emulator in the browser.
-   **WASI (WebAssembly System Interface)**: Provides the WASM applications with access to a virtualized filesystem and networking.

### 3. Unified Runner (pum-admin)
The `pum-admin` native Go application acts as the host. It:
-   Serves the Apptron assets and WASM binaries.
-   Injects necessary authentication and project metadata.
-   Bridges WASM applications to the backend Go microservices.

## Integration with the Go Frontend
Apptron doesn't replace the standard web experience; it enhances it. The Go-based frontend (`services/frontend`) remains the entry point for standard web access, while Apptron is the host for advanced management workflows and local-first execution.

## Note on x86 Emulation
While the Apptron platform possesses the technical capability for x86 emulation, this is **not** part of the current operational vision. Our focus remains on the high performance and portability of **Go-WASM/WASI** applications.
