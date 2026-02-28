# Migration Status and Roadmap

## Current Porting Status

The transition from PUM-Django to PUM-Go is an ongoing process. Below is the current status of the core application modules.

| Module | Status | Technology | Notes |
| :--- | :--- | :--- | :--- |
| **Identity / Accounts** | Ported | Go Microservice | Supports session management and auth. |
| **Product / PDU** | Ported | Go Microservice | Basic inventory and mock sync implemented. |
| **Inventory / GLPI** | Partial | Go Microservice | Integration engine in progress. |
| **Network / DNS / DHCP**| Pending | Django (Legacy) | Still running in the original monolith. |
| **GWS / VXLAN** | Pending | Django (Legacy) | Coordination logic needs porting to Go. |
| **WebSSH** | In Transition| WASM | Moving from server-side proxy to client-side WASM. |
| **Apptron Runner** | Active | Go (Native) | `pum-admin` is the primary host for the new UI. |

## The Migration Roadmap

### Phase 1: Foundation (Current)
-   Establish the Go microservices for Identity and Product.
-   Implement the Apptron Unified Runner (`pum-admin`).
-   Deliver the initial WASM-based `pum-cli`.
-   Create the mock-data layer for local development and testing.

### Phase 2: Core Service Porting
-   Port the **Network** and **Inventory** services from Django to Go.
-   Implement NATS or a similar lightweight bus for inter-service communication.
-   Expand the Action Registry to cover core networking tasks.

### Phase 3: Rich Client Enhancement
-   **WebSSH Port**: Complete the transition of WebSSH to a fully client-side WASM implementation within Apptron.
-   **TUI Deepening**: Build out comprehensive TUI wizards for device onboarding and configuration.
-   **Advanced Dashboards**: Develop rich WebViews for network topology and real-time analytics.

### Phase 4: Full Convergence
-   Retire the legacy Django monolith.
-   Implement production-grade service discovery and monitoring.
-   Deliver a unified "Command Center" installer that packages all microservices and the Apptron runner.

## The Goal
The end state is a system where the legacy Django code is entirely replaced by a suite of Go microservices, and the user interface is a high-performance Apptron application that feels native, responsive, and powerful.
