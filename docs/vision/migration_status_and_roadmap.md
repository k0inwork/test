# Migration Status and Roadmap

## Current Porting Status

The transition from PUM-Django to PUM-Go is an ongoing process. Below is the current status of the core application modules and frontends.

| Module | Status | Technology | Notes |
| :--- | :--- | :--- | :--- |
| **Identity / Accounts** | Ported | Go Microservice | Supports session management and auth. |
| **Product / PDU** | Ported | Go Microservice | Basic inventory and mock sync implemented. |
| **Inventory / GLPI** | Partial | Go Microservice | Integration engine in progress. |
| **Network / DNS / DHCP**| Pending | Django (Legacy) | Still running in the original monolith. |
| **GWS / VXLAN** | Pending | Django (Legacy) | Coordination logic needs porting to Go. |
| **Web Frontend** | Active | Go (Gin SSR) | Provides standard web access to core data. |
| **Apptron Runner** | Active | Go (Native) | Host for the new "Command Center" UI. |
| **WebSSH** | In Transition| WASM | Moving to a client-side WASM implementation. |

## The Migration Roadmap

### Phase 1: Foundation (Current)
-   Establish the Go microservices for Identity and Product.
-   Implement the Go-based web frontend for standard access.
-   Deliver the initial Apptron Unified Runner (`pum-admin`) and `pum-cli` (WASM).

### Phase 2: Core Service Porting
-   Port the **Network** and **Inventory** services from Django to Go.
-   Implement NATS or a similar lightweight bus for inter-service communication.
-   Expand the Action Registry to cover core networking tasks.

### Phase 3: Rich Client Enhancement
-   **WebSSH Port**: Complete the transition of WebSSH to a fully client-side WASM implementation within Apptron.
-   **TUI Deepening**: Build out comprehensive TUI wizards for device configuration.
-   **Advanced Dashboards**: Develop rich WebViews for network topology and analytics.

### Phase 4: Full Convergence
-   Retire the legacy Django monolith.
-   Deliver a unified "Command Center" installer packaging all microservices and the Apptron runner.

## The Goal
The end state is a system where the legacy Django code is replaced by Go microservices, accessible via a lightweight web frontend or a powerful, local-first Apptron "Command Center".
