# Unified UI Strategy: Web, GUI, CLI, and TUI

## The Philosophy of Choice
Network administrators have diverse preferences and access requirements. Our vision is to support a spectrum of interfaces—from standard web templates to high-performance WASM tools—ensuring that every user has the right tool for the task at hand.

## Interface Tiers

### 1. Standard Web Access (SSR)
-   **Purpose**: Lightweight, zero-dependency access to core data and status.
-   **Technology**: Go-based Gin templates (`services/frontend`).
-   **Best For**: Quick checks, mobile access, and standard reporting.

### 2. Dashboards (Apptron WebView)
-   **Purpose**: Deep monitoring, discovery, and complex data visualization.
-   **Technology**: Apptron WebViews (HTML/JS/CSS).
-   **Best For**: Seeing "The Big Picture" and interactive topology management.

### 3. CLI (WASM)
-   **Purpose**: Scripting, automation, and quick one-off commands.
-   **Technology**: Go-WASM / Cobra running in Apptron.
-   **Best For**: "Power Users" and batch operations.

### 4. TUI (WASM)
-   **Purpose**: Interactive configuration and focused management tasks.
-   **Technology**: Go-WASM / Bubbletea running in Apptron.
-   **Best For**: Complex workflows that require step-by-step guidance but benefit from keyboard navigation.

## The Action Registry Pattern
To achieve consistency across these four tiers without duplicating code, we employ the **Action Registry** pattern.

1.  **Core Action**: A single Go struct defines the logic for an operation.
2.  **Validation**: Shared validation rules ensure consistency.
3.  **Dispatch**:
    -   **Web/GUI**: A button click triggers the Action via a REST/GraphQL call.
    -   **CLI/TUI**: The Action logic is executed directly within the WASM environment, leveraging local processing power.

## Unified Experience
The Apptron environment serves as the ultimate "Command Center" where these interfaces (except for the standalone Web SSR) converge, providing a seamless transition between visual dashboards and keyboard-driven management tools.
