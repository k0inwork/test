# Migration Strategy Proposal v2: Bridge to Convergence

## Executive Summary
This proposal outlines the next steps for the PUM migration from Django to Go. The focus shifts from pure feature porting to ensuring **external system stability** and **architectural continuity**. We propose the introduction of a dedicated **Compatibility Service** that bridges the gap between the new Go microservices and legacy external dependencies (GUIs, older management tools).

---

## Strategic Goals
1.  **Non-Disruptive Migration**: Ensure external systems (GUIs) continue to function using original endpoints.
2.  **API Fidelity**: Mimic legacy Django JSON output formats, including the `json=true` middleware behavior.
3.  **Infrastructure Continuity**: Maintain support for legacy RabbitMQ message formats (JSON/YAML) to ensure inter-module communication remains intact.
4.  **Data Parity**: Replicate legacy SQL views (`ports_view`, `switch_view`) once the `inventory` service is fully operational.

---

## Option 1: Compatibility Layer Focus (Recommended)
**Prioritize the creation of a standalone `compatibility` service.**

-   **Description**: A dedicated service that routes `compatibility/<original_url>` to the new Go services.
-   **Key Features**:
    -   Implementation of the `json=true` logic (context interception and serialization).
    -   Manual endpoint mapping (e.g., `/compatibility/users` -> `identity:8081/users`).
    -   Isolation of legacy support from the core Go service logic.
-   **Pros**: High stability for external systems; clean core services.
-   **Cons**: Initial overhead in implementing the mapping service.

## Option 2: Accelerated Module Porting
**Focus on porting `network` and `inventory` while building compatibility into the porting process.**

-   **Description**: Parallelize the porting of remaining Django modules with the development of the compatibility layer.
-   - **Key Features**:
    -   Immediate porting of `network` (Phase 2).
    -   Development of `inventory` with integrated support for `ports_view` and `switch_view`.
-   **Pros**: Faster decommissioning of the Django monolith.
-   **Cons**: Higher complexity and risk of regressions in external system integrations.

## Option 3: Infrastructure Convergence
**Focus on the messaging and data synchronization backbone.**

-   **Description**: Prioritize the `pkg/messaging` layer and the `data-sync-engine` port.
-   **Key Features**:
    -   Formalize Go-based RabbitMQ handlers that match Python `event_consumer` signatures.
    -   Migrate the rename-detection logic (Redis-based) to Go.
-   **Pros**: Solidifies the distributed architecture foundation.
-   **Cons**: Delayed visible progress on API and UI parity.

---

## Technical Implementation Roadmap

### 1. The Compatibility Service (`/services/compatibility`)
This service will act as a "Translation Gateway":
-   **Input**: HTTP request to `/compatibility/v1/products/?json=true`
-   **Process**:
    1. Fetch raw data from `services/product`.
    2. Apply "Legacy Serialization" rules (matching the Python `jsonify` engine).
    3. Return a response structure identical to the Django 4.0 context.

### 2. Legacy SQL Views (Phase: Inventory Readiness)
Once the `inventory` service is stable, it will be responsible for providing:
-   `ports_view`: Joined view of `switch_ports`, `switches`, and `nodes`.
-   `switch_view`: Detailed hardware state view.
-   **Note**: These views will be created via Go-based migration scripts using `gorm.Exec`.

### 3. Messaging Bridge (`/pkg/messaging`)
A shared library implementing:
-   YAML/JSON serialization compatible with Python's `ip-module` and `switch-module`.
-   Consistent header and routing key patterns.

---

## Next Steps
1.  Approve **Option 1** as the immediate path forward.
2.  Implement the prototype `compatibility` service for `identity` and `product` endpoints.
3.  Establish the `pkg/messaging` baseline.
