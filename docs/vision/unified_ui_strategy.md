# Unified UI Strategy: GUI, CLI, and TUI

## The Philosophy of Choice
Network administrators have diverse preferences. Some prefer the visual clarity of a GUI, while others demand the speed of a CLI or the focused interactivity of a TUI. Our vision is to support all three as first-class citizens, ensuring no functionality is lost regardless of the interface chosen.

## The Action Registry Pattern
To achieve this without duplicating code, we employ the **Action Registry** pattern.

1.  **Core Action**: A single Go struct defines the logic for an operation (e.g., "Reset Port").
2.  **Validation**: Shared validation rules ensure consistency.
3.  **Dispatch**:
    -   **GUI (WebView)**: A button click triggers the Action via a REST/GraphQL call to the bridge.
    -   **CLI (WASM)**: A command line argument (e.g., `pum port reset`) executes the Action logic directly in WASM.
    -   **TUI (WASM)**: Selecting an item from a list in the Bubbletea interface triggers the same Action logic.

## Interface Roles

### Dashboards (GUI)
-   **Purpose**: Monitoring, discovery, and complex data visualization.
-   **Technology**: WebViews (HTML/JS/CSS) served via the unified runner.
-   **Best For**: Seeing "The Big Picture".

### CLI (WASM)
-   **Purpose**: Scripting, automation, and quick one-off commands.
-   **Technology**: Go-WASM / Cobra.
-   **Best For**: "Power Users" and batch operations.

### TUI (WASM)
-   **Purpose**: Interactive configuration and focused management tasks.
-   **Technology**: Go-WASM / Bubbletea.
-   **Best For**: Complex workflows that require step-by-step guidance but benefit from keyboard navigation.

## Terminal Integration
The Apptron environment includes a built-in terminal that hosts the WASM-based CLI and TUI. This creates a seamless workflow where an admin can check a dashboard and immediately drop into a TUI to perform a remediation, all within the same application shell.
