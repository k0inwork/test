# Migration Status and Roadmap

## Current Porting Status

The transition from PUM-Django to PUM-Go is an ongoing process. Below is the current status of the core application modules and frontends.

| Module | Status | Technology | Notes |
| :--- | :--- | :--- | :--- |
| **Identity / Accounts** | Ported | Go Microservice | Supports session management and auth. |
| **Product / PDU** | Ported | Go Microservice | Basic inventory and mock sync implemented. |
| **Inventory / GLPI** | Partial | Go Microservice | Integration engine in progress. |
| **Network / DNS / DHCP**| Pending | Django (Legacy) | High priority for Phase 2 porting. |
| **GWS** | Pending | Django (Legacy) | Logical management points to be ported. |
| **Services / Connectivity**| Pending | Django (Legacy) | Key/Data services to be ported in Phase 3. |
| **Integrity** | Pending | Django (Legacy) | Validation rules to be ported in Phase 3. |
| **Web Frontend** | Active | Go (Gin SSR) | Provides standard web access to core data. |
| **Apptron Runner** | Active | Go (Native) | Host for the new "Command Center" UI. |
| **WebSSH** | In Transition| WASM | Moving to a client-side WASM implementation. |

## The Migration Roadmap (The Progress Line)

### Phase 1: Foundation (Current)
-   Establish the Go microservices for Identity and Product.
-   Implement the Go-based web frontend for standard access.
-   Deliver the initial Apptron Unified Runner (`pum-admin`) and `pum-cli` (WASM).
-   **Technical Milestone**: Implementation of the Action Registry for initial node management.

### Phase 2: Core Network & Inventory
-   Port the **Network** and **Inventory** services from Django to Go.
-   **Technical Milestone**: Transition inter-service communication to a high-performance bus (e.g., NATS).
-   Expand the Action Registry to cover core networking tasks (Port management, VLANs).
-   Implement the initial Go-based "Hardware Proxies" for Switches.

### Phase 3: Service Abstraction & Rich Client
-   Port the **Services** and **Integrity** modules to Go.
-   **Technical Milestone**: Complete transition of WebSSH to a fully client-side WASM implementation within Apptron.
-   Migration of the **Data Sync Engine** to the new architecture.
-   Deliver rich Apptron WebViews for network topology and analytics.

### Phase 4: Full Convergence & Infrastructure
-   Migration of "Multi-Master Coordination" to a Go-native implementation.
-   Full decommissioning of the legacy Django monolith and RabbitMQ.
-   Deliver a unified "Command Center" installer packaging all microservices.

## The Goal
The end state is a system where the legacy Django code is entirely replaced by a suite of high-performance Go microservices, and the user interface is a unified, capability-driven Apptron application.
