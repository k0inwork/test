# Apptron Integration Technical Design

## Architecture Overview

The system consists of three main components working together within the Apptron environment:

1.  **Headless WASM Core**: A Go service compiled to WASM running in a WebWorker. It acts as the local data plane, caching state from the main microservices and providing a local API for the UI.
2.  **CLI/TUI Tool (pum-cli)**: A Go binary compiled to WASM that runs inside Apptron's emulated Alpine Linux. It detects its environment and switches between CLI (scripting) and TUI (interactive) modes.
3.  **Web View Dashboards**: Standard HTML/JS dashboards that communicate with the WASM Core via the Apptron ServiceWorker or a SharedWorker.
4.  **Bridge Agent (The Gateway)**: A native Go application running on the admin's local network that bridges the Apptron virtual network to the physical management subnet.

## The Bridge Agent (Local Gateway)

To reach physical network devices from the browser, we need a bridge.

- **Mechanism**: The Bridge Agent connects to the Apptron Gateway (via WebSockets) and joins the session's virtual network.
- **Routing**: It acts as a Layer 3 gateway. Traffic from the Apptron browser environment destined for the management subnet is tunneled through the WebSocket to the Bridge Agent, which then forwards it to the physical devices.
- **Security**: Uses mTLS or Session Tokens to authorize the Bridge Agent.

## WASM Execution Modes

### 1. In-Browser WASM Server
- **Tech**: Go + syscall/js + WebWorkers.
- **Role**: Provides a local REST/GraphQL API for Web Views.
- **Benefit**: Extremely low latency for UI updates; works even if the main server is slow.

### 2. In-Alpine WASM Binary (Wanix)
- **Tech**: Go + WASI (via Wanix).
- **Role**: The `pum-cli` tool.
- **Workflow**:
  - Run `pum status` for quick CLI info.
  - Run `pum dashboard` to launch the TUI.

## Integration Strategy

1.  **Subcloning**: We will include Apptron as a submodule or subfolder in the main repo.
2.  **Custom Bundles**: We will create a custom `sys.tar.gz` bundle for Apptron that includes the `pum-cli` pre-installed in `/bin`.
3.  **Environment Variables**: The Apptron environment will be pre-configured with `PUM_SERVER_URL` pointing to the main Go microservices.

## Communication Flow

[Dashboard (Web)] <-> [WASM Core (Worker)] <-> [Bridge Agent (Native)] <-> [Network Device]
                               |
                        [PUM-Go Microservices]

## Implementation Roadmap

### Phase 1: Prototype WASM Client
- Implement basic `pum-cli` in Go.
- Compile to WASM using `GOOS=js GOARCH=wasm`.
- Test running in a standard Apptron instance via Wanix.

### Phase 2: Bridge Agent Development
- Create the `pum-bridge` agent.
- Implement WebSocket tunneling to a mock Apptron Gateway.
- Verify L3 routing from browser to local mock device.

### Phase 3: Web Dashboard & Headless Core
- Develop the "Headless Core" in Go-WASM.
- Create a sample HTML/JS dashboard using `html-component`.
- Establish communication between Dashboard and Headless Core via ServiceWorker.

### Phase 4: Apptron Customization
- Fork Apptron (if needed) or create a custom build script.
- Pre-install PUM tools in the default image.
- Customize the VSCode workbench layout for network admins.
