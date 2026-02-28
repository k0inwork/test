# Transition Strategy: The Evolution from Django to Go

## The "Strangler Fig" Approach
We are not attempting a "Big Bang" rewrite. Instead, we are using the **Strangler Fig** pattern to gradually replace functionality within the legacy Django monolith with modern Go microservices.

### 1. Unified Entry Point
The Go-based Frontend and the Apptron Unified Runner serve as the new face of the system.
-   **New Features**: Built exclusively as Go microservices and integrated into the Apptron UI via WASM/WebViews.
-   **Legacy Features**: Initially surfaced in the new UI by proxying to the legacy Django views or utilizing the existing Django JSON APIs.

### 2. Incremental Porting
Functionality is moved from Django to Go module-by-module, following a clear priority:
-   **Foundation (Done)**: Identity, Session Management, Basic Product Inventory.
-   **Core Network (High Priority)**: Network allocation, DNS/DHCP integration, and GWS logical points.
-   **Complex Systems (Medium Priority)**: Data Sync Engine, Service/Connectivity management, and Integrity rules.
-   **Hardware Proxies (Ongoing)**: Porting the Python-based hardware modules (PDU, Switch, IPMI) to Go-based providers.

## Managing Coexistence
During the transition, the system operates in a hybrid state.

### Communication Bridge
-   The legacy Django monolith and the new Go microservices will temporarily coexist on the same network.
-   We utilize the **Registry Service** to allow Go services to discover one another, while potentially bridging to RabbitMQ for communication with unported legacy modules.

### Shared State and Data Migration
-   **Database Transition**: As modules are ported, their data models are migrated from the monolithic PostgreSQL database to service-specific schemas or new databases.
-   **State Consistency**: During the migration of core systems like the "Data Sync Engine", we implement dual-writing or synchronization tasks to ensure both the legacy and new systems have a consistent view of the network.

## High-Level Progress Line

### Phase 1: The Shell (Current)
-   Go-based Frontend and Apptron Runner are operational.
-   Identity and Product services are the first to be fully ported.
-   Initial WASM tools (`pum-cli`) are available.

### Phase 2: Core Network & Inventory
-   Porting of Switch and Port management.
-   Porting of Network and GWS logical points.
-   Full implementation of the Action Registry for core networking tasks.

### Phase 3: Service Abstraction & Automation
-   Porting of "Services" (Connectivity/Data/Key services).
-   Migration of the Data Sync Engine to Go.
-   Transition of WebSSH to full client-side WASM.

### Phase 4: Hardware & Infrastructure
-   Migration of the "Multi-Master Coordination" logic.
-   Porting of remaining hardware controller modules.
-   Decommissioning of the Django monolith and RabbitMQ.

## The Goal
A clean, high-performance, and decentralized ecosystem where every management task is a Go-powered Action, delivered through a rich, local-first Apptron experience.
